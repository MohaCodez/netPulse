import './Theory.css';

export function Theory() {
  return (
    <div className="theory-page">
      <h2>📚 Network Theory Reference</h2>
      <p className="theory-intro">
        Concepts behind each chart — what to look for, why it matters, and how to interpret anomalies.
      </p>

      <section className="theory-section">
        <h3>📶 Wi-Fi Signal Strength (RSSI)</h3>
        <div className="theory-content">
          <p>
            <strong>RSSI (Received Signal Strength Indicator)</strong> measures the power level of the received radio signal in dBm.
            It's a logarithmic scale — every 3 dB change represents a doubling/halving of power.
          </p>
          <table className="theory-table">
            <thead><tr><th>Range</th><th>Quality</th><th>Expected Behavior</th></tr></thead>
            <tbody>
              <tr><td>-30 to -50 dBm</td><td>Excellent</td><td>Max throughput, minimal retransmissions</td></tr>
              <tr><td>-50 to -60 dBm</td><td>Good</td><td>Reliable, minor throughput reduction</td></tr>
              <tr><td>-60 to -70 dBm</td><td>Fair</td><td>Noticeable throughput drop, occasional retransmits</td></tr>
              <tr><td>-70 to -80 dBm</td><td>Poor</td><td>High retransmission rate, latency spikes</td></tr>
              <tr><td>Below -80 dBm</td><td>Unusable</td><td>Frequent disconnects, massive packet loss</td></tr>
            </tbody>
          </table>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>Free-space path loss:</strong> Signal attenuates with distance² (inverse square law). Each wall adds ~3-6 dB attenuation depending on material.</li>
              <li><strong>Multipath fading:</strong> Reflected signals arrive at different phases, causing constructive/destructive interference. This is why signal fluctuates even when stationary.</li>
              <li><strong>SNR (Signal-to-Noise Ratio):</strong> What matters isn't absolute signal, but signal relative to noise floor. SNR = Signal - Noise (in dBm). Need &gt;20 dB for reliable data.</li>
              <li><strong>Shannon's theorem:</strong> Channel capacity C = B × log₂(1 + SNR). Lower SNR → lower max throughput regardless of link speed.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>🔄 Band Hopping (2.4 GHz vs 5 GHz)</h3>
        <div className="theory-content">
          <p>
            Modern routers use <strong>band steering</strong> to move clients between 2.4 GHz and 5 GHz. Each band has trade-offs:
          </p>
          <table className="theory-table">
            <thead><tr><th>Property</th><th>2.4 GHz</th><th>5 GHz</th></tr></thead>
            <tbody>
              <tr><td>Range</td><td>Longer (~45m indoor)</td><td>Shorter (~15m indoor)</td></tr>
              <tr><td>Throughput</td><td>Lower (max ~150 Mbps)</td><td>Higher (max ~1300 Mbps)</td></tr>
              <tr><td>Channels</td><td>3 non-overlapping (1, 6, 11)</td><td>24 non-overlapping</td></tr>
              <tr><td>Interference</td><td>High (microwaves, Bluetooth, neighbors)</td><td>Low (fewer devices)</td></tr>
              <tr><td>Wall penetration</td><td>Good</td><td>Poor</td></tr>
            </tbody>
          </table>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>Wavelength vs penetration:</strong> λ = c/f. 2.4 GHz → λ≈12.5cm, 5 GHz → λ≈6cm. Longer wavelengths diffract around obstacles better.</li>
              <li><strong>Channel bandwidth:</strong> 2.4 GHz channels are 20 MHz wide (40 MHz bonded). 5 GHz supports 80/160 MHz — directly increasing throughput.</li>
              <li><strong>Band steering hysteresis:</strong> Routers use signal thresholds with hysteresis to avoid rapid switching. If you see frequent hops, the thresholds are poorly tuned.</li>
              <li><strong>Roaming decision:</strong> The client (not the AP) ultimately decides when to switch. Linux's wpa_supplicant uses `bss_transition` and signal delta thresholds.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>🔀 Gateway vs External Latency</h3>
        <div className="theory-content">
          <p>
            By comparing latency to your router (1 hop) vs external servers (10+ hops), you isolate <strong>where</strong> delay occurs.
          </p>
          <div className="theory-key-points">
            <h4>Interpretation:</h4>
            <ul>
              <li><strong>Gateway high, external high:</strong> Local problem (Wi-Fi interference, router overload). Fix locally.</li>
              <li><strong>Gateway low, external high:</strong> ISP problem (congestion, routing). Nothing you can fix.</li>
              <li><strong>Both spike together:</strong> Likely a shared bottleneck (bufferbloat at router, WAN link saturated).</li>
              <li><strong>Gateway normal is 1-5ms.</strong> If consistently &gt;20ms, your Wi-Fi link is struggling.</li>
            </ul>
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>Bufferbloat:</strong> When router buffers are too large, packets queue up instead of being dropped. Causes latency without packet loss. Detected by high latency + low loss.</li>
              <li><strong>CSMA/CA contention:</strong> On Wi-Fi, devices compete for airtime. More devices = more backoff = higher gateway latency even with strong signal.</li>
              <li><strong>Hop-by-hop analysis:</strong> Each network hop adds processing delay (store-and-forward) + propagation delay (speed of light in fiber ≈ 200km/ms).</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>📊 Signal ↔ Latency Correlation</h3>
        <div className="theory-content">
          <p>
            This chart overlays Wi-Fi signal and gateway latency on the same time axis.
            When they move inversely (signal drops → latency rises), it proves <strong>causal relationship</strong> between signal quality and network performance.
          </p>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>PHY rate adaptation:</strong> Wi-Fi adapts modulation scheme (MCS index) based on signal quality. Lower signal → lower MCS → more airtime per packet → higher latency for everyone.</li>
              <li><strong>Retransmission overhead:</strong> At low signal, frame error rate increases. Each retransmit adds ~10ms. At -75 dBm, retransmit rates can exceed 30%.</li>
              <li><strong>Pearson correlation coefficient:</strong> Calculate r between signal and latency time series. r &lt; -0.5 indicates strong inverse relationship (expected for Wi-Fi bottleneck).</li>
              <li><strong>Granger causality:</strong> For research, test if past signal values predict future latency better than past latency alone. This establishes directional causation.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>📉 Packet Loss</h3>
        <div className="theory-content">
          <p>
            Packet loss occurs when a transmitted packet never arrives at the destination. On Wi-Fi, it's usually due to interference or signal weakness. On wired links, it indicates congestion or hardware failure.
          </p>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>Random vs bursty loss:</strong> Random loss (uniform distribution) = congestion. Bursty loss (clusters) = interference or fading. Check inter-loss intervals.</li>
              <li><strong>TCP behavior under loss:</strong> TCP halves its congestion window on each loss event. Even 1% loss can reduce throughput by 50%+ due to exponential backoff.</li>
              <li><strong>Hidden terminal problem:</strong> Device A can't hear Device B, both transmit to AP simultaneously → collision → loss. Solved by RTS/CTS but adds overhead.</li>
              <li><strong>FEC (Forward Error Correction):</strong> Modern Wi-Fi (802.11ax) uses LDPC codes that can correct some bit errors without retransmission. Loss only occurs when FEC capacity is exceeded.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>〰️ Jitter</h3>
        <div className="theory-content">
          <p>
            Jitter is the <strong>variance in latency</strong> over time. Constant 50ms latency is fine for VoIP; 50ms ± 40ms jitter is unusable. Measured as standard deviation of inter-packet delay.
          </p>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>De-jitter buffers:</strong> VoIP/video apps use buffers to smooth jitter. Buffer size = trade-off between smoothness and added latency. Typically 30-50ms.</li>
              <li><strong>MOS (Mean Opinion Score):</strong> Voice quality metric. Jitter &gt;30ms typically drops MOS below 3.5 (unacceptable). ITU-T G.114 recommends &lt;150ms one-way delay including jitter buffer.</li>
              <li><strong>QoS (Quality of Service):</strong> Jitter is caused by variable queuing delays. WMM (Wi-Fi Multimedia) prioritizes voice/video traffic to reduce their jitter.</li>
              <li><strong>Measurement:</strong> Jitter = √(Σ(latency_i - mean)² / N). We compute this per ping burst (3 packets) as the standard deviation of RTT.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>🌐 DNS Resolution</h3>
        <div className="theory-content">
          <p>
            DNS translates domain names to IP addresses. Resolution time directly impacts page load — every new domain requires a lookup before the first byte can be fetched.
          </p>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>Recursive vs iterative resolution:</strong> Your resolver recursively queries root → TLD → authoritative servers. Each hop adds network RTT.</li>
              <li><strong>Cache hierarchy:</strong> Browser cache → OS cache → resolver cache → upstream. Cold cache = ~100ms. Warm cache = ~1ms. TTL controls cache duration.</li>
              <li><strong>Anycast routing:</strong> Google (8.8.8.8) and Cloudflare (1.1.1.1) use anycast — same IP, multiple physical locations. Nearest one responds. Performance varies by ISP peering.</li>
              <li><strong>DNS-over-HTTPS (DoH):</strong> Encrypts DNS to prevent snooping/tampering but adds TLS handshake overhead (~50ms first query). Subsequent queries use persistent connections.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>🗓️ Reliability Heatmap</h3>
        <div className="theory-content">
          <p>
            The heatmap shows probe success rate by hour of day over 7 days. Patterns reveal systematic issues tied to time-of-use.
          </p>
          <div className="theory-key-points">
            <h4>What patterns mean:</h4>
            <ul>
              <li><strong>Evening degradation (6-11 PM):</strong> ISP congestion — shared neighborhood bandwidth saturated during peak hours. Common with cable/GPON.</li>
              <li><strong>Night improvement (2-6 AM):</strong> Confirms ISP congestion hypothesis — same infrastructure, fewer users.</li>
              <li><strong>Random scattered failures:</strong> Interference (microwave ovens, baby monitors on 2.4 GHz) or hardware instability.</li>
              <li><strong>Consistent failures at specific hours:</strong> Scheduled processes (backups, updates) consuming bandwidth. Check what else runs on your network.</li>
            </ul>
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>Statistical confidence:</strong> Need &gt;100 samples per cell for reliable success rate. Low sample counts show noisy colors.</li>
              <li><strong>GPON contention ratio:</strong> Fiber ISPs typically share 1:32 or 1:64 split ratio. Your 2.5 Gbps GPON port is shared with 32-64 households.</li>
            </ul>
          </div>
        </div>
      </section>

      <section className="theory-section">
        <h3>🚀 Speed Tests</h3>
        <div className="theory-content">
          <p>
            Speed tests measure maximum achievable throughput by saturating the link with parallel TCP connections.
          </p>
          <div className="theory-key-points">
            <h4>Key ECE Concepts:</h4>
            <ul>
              <li><strong>TCP window scaling:</strong> Throughput ≤ WindowSize / RTT. With 200ms RTT, need 5MB window for 200 Mbps. Multiple connections work around this.</li>
              <li><strong>Bandwidth-delay product (BDP):</strong> BDP = bandwidth × RTT. Buffers must be ≥ BDP to avoid underutilization. Our 6 parallel connections approximate this.</li>
              <li><strong>Last-mile bottleneck:</strong> Speed is limited by the weakest link. Wi-Fi PHY rate &gt; ISP plan? ISP is bottleneck. ISP plan &gt; Wi-Fi? Wi-Fi is bottleneck.</li>
              <li><strong>Why speed varies:</strong> Server load, ISP throttling (DPI-based), TCP congestion control algorithm (CUBIC vs BBR), time-of-day congestion.</li>
            </ul>
          </div>
        </div>
      </section>
    </div>
  );
}
