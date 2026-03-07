#![no_std]
#![no_main]
#![feature(impl_trait_in_assoc_type)]

use cyw43::JoinOptions;
use cyw43_pio::PioSpi;
use embassy_executor::Spawner;
use embassy_futures::select::{select, Either};
use embassy_net::{Stack, StackResources};
use embassy_rp::{
    adc::{Adc, Async as AdcAsync, Channel as AdcChannel, Config as AdcConfig},
    gpio::{Level, Output},
    peripherals::{DMA_CH0, PIO0},
    pio::Pio,
};
use embassy_sync::{blocking_mutex::raw::ThreadModeRawMutex, channel::Channel, mutex::Mutex};
use embassy_time::Timer;
use picoserve::{make_static, response::Json, routing::get, AppBuilder, AppRouter};
use rand::Rng;
use serde::Serialize;

use panic_persist as _;

mod secrets;
use secrets::{WIFI_PASSWORD, WIFI_SSID};

// ---------------------------------------------------------------------------
// Interrupt bindings
// ---------------------------------------------------------------------------
embassy_rp::bind_interrupts!(struct Irqs {
    PIO0_IRQ_0   => embassy_rp::pio::InterruptHandler<embassy_rp::peripherals::PIO0>;
    USBCTRL_IRQ  => embassy_rp::usb::InterruptHandler<embassy_rp::peripherals::USB>;
    ADC_IRQ_FIFO => embassy_rp::adc::InterruptHandler;
});

// ---------------------------------------------------------------------------
// Shared globals
// ---------------------------------------------------------------------------

/// Non-blocking channel: web handlers send blink counts; main loop drives LED.
/// Capacity 4 — drop silently if full rather than blocking request handling.
static BLINK_CHANNEL: Channel<ThreadModeRawMutex, u32, 4> = Channel::new();

/// ADC + temp-sensor channel stored together so a single lock covers both.
/// ThreadModeRawMutex is safe on single-core RP2040 with cooperative scheduling.
struct AdcState {
    adc: Adc<'static, AdcAsync>,
    channel: AdcChannel<'static>,
}

static ADC_STATE: Mutex<ThreadModeRawMutex, Option<AdcState>> = Mutex::new(None);

// ---------------------------------------------------------------------------
// Task: USB serial logger
// ---------------------------------------------------------------------------
#[embassy_executor::task]
async fn logger_task(usb: embassy_rp::Peri<'static, embassy_rp::peripherals::USB>) {
    let driver = embassy_rp::usb::Driver::new(usb, Irqs);
    embassy_usb_logger::run!(1024, log::LevelFilter::Info, driver);
}

// ---------------------------------------------------------------------------
// Task: CYW43 radio runner
// ---------------------------------------------------------------------------
#[embassy_executor::task]
async fn wifi_task(
    runner: cyw43::Runner<'static, Output<'static>, PioSpi<'static, PIO0, 0, DMA_CH0>>,
) -> ! {
    runner.run().await
}

// ---------------------------------------------------------------------------
// Task: embassy-net stack runner
// ---------------------------------------------------------------------------
#[embassy_executor::task]
async fn net_task(mut runner: embassy_net::Runner<'static, cyw43::NetDriver<'static>>) -> ! {
    runner.run().await
}

// ---------------------------------------------------------------------------
// Temperature reading
// ---------------------------------------------------------------------------

#[derive(Serialize)]
struct TempReading {
    #[serde(rename = "tempC")]
    temp_c: f32,
    #[serde(rename = "tempF")]
    temp_f: f32,
}

async fn read_temperature() -> TempReading {
    let mut guard = ADC_STATE.lock().await;
    if let Some(state) = guard.as_mut() {
        let raw: u16 = state.adc.read(&mut state.channel).await.unwrap_or(0);
        // RP2040 datasheet formula: Vref = 3.3V, 12-bit ADC
        let voltage = (raw as f32) * 3.3_f32 / 4096.0_f32;
        let temp_c = 27.0_f32 - (voltage - 0.706_f32) / 0.001721_f32;
        TempReading {
            temp_c,
            temp_f: temp_c * 9.0_f32 / 5.0_f32 + 32.0_f32,
        }
    } else {
        TempReading {
            temp_c: 0.0,
            temp_f: 32.0,
        }
    }
}

// ---------------------------------------------------------------------------
// HTTP application
// ---------------------------------------------------------------------------

struct AppProps;

impl AppBuilder for AppProps {
    type PathRouter = impl picoserve::routing::PathRouter;

    fn build_app(self) -> picoserve::Router<Self::PathRouter> {
        picoserve::Router::new().route(
            "/",
            get(|| async move {
                let reading = read_temperature().await;
                // Signal 5 blinks; non-blocking — drop if channel full
                let _ = BLINK_CHANNEL.try_send(5);
                log::info!("temp: {:.2}C / {:.2}F", reading.temp_c, reading.temp_f);
                Json(reading)
            }),
        )
    }
}

static CONFIG: picoserve::Config = picoserve::Config::const_default();

// 8 concurrent web tasks — each handles one connection at a time
const WEB_TASK_POOL_SIZE: usize = 8;

// ---------------------------------------------------------------------------
// Task: HTTP worker (one per concurrent connection slot)
// ---------------------------------------------------------------------------
#[embassy_executor::task(pool_size = WEB_TASK_POOL_SIZE)]
async fn web_task(
    task_id: usize,
    stack: Stack<'static>,
    app: &'static AppRouter<AppProps>,
) -> ! {
    let port = 80;
    let mut tcp_rx_buffer = [0u8; 1024];
    let mut tcp_tx_buffer = [0u8; 1024];
    let mut http_buffer = [0u8; 2048];

    picoserve::Server::new(app, &CONFIG, &mut http_buffer)
        .listen_and_serve(task_id, stack, port, &mut tcp_rx_buffer, &mut tcp_tx_buffer)
        .await
        .into_never()
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------
#[embassy_executor::main]
async fn main(spawner: Spawner) {
    let p = embassy_rp::init(Default::default());

    spawner.must_spawn(logger_task(p.USB));

    // -- ADC: RP2040 internal temperature sensor ----------------------------
    // Both peripherals originate from embassy_rp::init() so they are 'static.
    {
        let adc = Adc::new(p.ADC, Irqs, AdcConfig::default());
        let channel = AdcChannel::new_temp_sensor(p.ADC_TEMP_SENSOR);
        *ADC_STATE.lock().await = Some(AdcState { adc, channel });
    }

    // -- CYW43439 WiFi chip -------------------------------------------------
    // Replace these placeholder blobs with real firmware before flashing.
    // Obtain from: https://github.com/embassy-rs/embassy/tree/main/cyw43-firmware
    let fw = include_bytes!("../cyw43-firmware/43439A0.bin");
    let clm = include_bytes!("../cyw43-firmware/43439A0_clm.bin");

    let pwr = Output::new(p.PIN_23, Level::Low);
    let cs = Output::new(p.PIN_25, Level::High);
    let mut pio = Pio::new(p.PIO0, Irqs);
    let spi = PioSpi::new(
        &mut pio.common,
        pio.sm0,
        cyw43_pio::DEFAULT_CLOCK_DIVIDER,
        pio.irq0,
        cs,
        p.PIN_24,
        p.PIN_29,
        p.DMA_CH0,
    );

    let cyw_state = make_static!(cyw43::State, cyw43::State::new());
    let (net_device, mut control, runner) = cyw43::new(cyw_state, pwr, spi, fw).await;
    spawner.must_spawn(wifi_task(runner));

    control.init(clm).await;
    control
        .set_power_management(cyw43::PowerManagementMode::PowerSave)
        .await;

    // -- embassy-net stack (DHCP) -------------------------------------------
    let seed: u64 = embassy_rp::clocks::RoscRng.random();

    let (stack, net_runner) = embassy_net::new(
        net_device,
        embassy_net::Config::dhcpv4(Default::default()),
        make_static!(
            StackResources<{ WEB_TASK_POOL_SIZE + 2 }>,
            StackResources::new()
        ),
        seed,
    );
    spawner.must_spawn(net_task(net_runner));

    // -- Join WiFi in STA mode ----------------------------------------------
    log::info!("Joining WiFi SSID: {}", WIFI_SSID);
    loop {
        match control
            .join(WIFI_SSID, JoinOptions::new(WIFI_PASSWORD.as_bytes()))
            .await
        {
            Ok(_) => {
                log::info!("WiFi joined");
                break;
            }
            Err(e) => {
                log::warn!("WiFi join failed: status={}", e.status);
                Timer::after_secs(5).await;
            }
        }
    }

    // Wait for DHCP lease before advertising readiness
    stack.wait_config_up().await;
    log::info!("Network up: {:?}", stack.config_v4());

    // -- Spawn HTTP workers -------------------------------------------------
    let app = make_static!(AppRouter<AppProps>, AppProps.build_app());
    for task_id in 0..WEB_TASK_POOL_SIZE {
        spawner.must_spawn(web_task(task_id, stack, app));
    }

    // -- Main loop: WiFi watchdog + LED blink controller --------------------
    // control stays here so we can both reconnect WiFi and drive the CYW43 LED.
    log::info!("Entering watchdog loop");

    let mut watchdog_deadline =
        embassy_time::Instant::now() + embassy_time::Duration::from_secs(30);

    loop {
        match select(BLINK_CHANNEL.receive(), Timer::at(watchdog_deadline)).await {
            Either::First(n) => {
                // Blink LED n times, ending with LED off
                for _ in 0..n {
                    control.gpio_set(0, true).await;
                    Timer::after_millis(250).await;
                    control.gpio_set(0, false).await;
                    Timer::after_millis(250).await;
                }
            }
            Either::Second(_) => {
                // Watchdog tick: check WiFi link, reconnect if needed
                watchdog_deadline =
                    embassy_time::Instant::now() + embassy_time::Duration::from_secs(30);

                if stack.config_v4().is_none() {
                    log::warn!("WiFi link lost — reconnecting...");
                    loop {
                        match control
                            .join(WIFI_SSID, JoinOptions::new(WIFI_PASSWORD.as_bytes()))
                            .await
                        {
                            Ok(_) => {
                                log::info!("WiFi rejoined");
                                break;
                            }
                            Err(e) => {
                                log::warn!("Reconnect failed: status={} — retrying in 5s", e.status);
                                Timer::after_secs(5).await;
                            }
                        }
                    }
                } else {
                    log::info!("heartbeat: wifi ok, ip={:?}", stack.config_v4());
                }
            }
        }
    }
}
