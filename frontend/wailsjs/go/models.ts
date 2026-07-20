export namespace main {
	
	export class AlertRules {
	    latency_warning_ms: number;
	    latency_critical_ms: number;
	    loss_warning_pct: number;
	    loss_critical_pct: number;
	    jitter_warning_ms: number;
	    wifi_signal_warning: number;
	    wifi_signal_critical: number;
	    notifications_enabled: boolean;
	    notification_cooldown_sec: number;
	
	    static createFrom(source: any = {}) {
	        return new AlertRules(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latency_warning_ms = source["latency_warning_ms"];
	        this.latency_critical_ms = source["latency_critical_ms"];
	        this.loss_warning_pct = source["loss_warning_pct"];
	        this.loss_critical_pct = source["loss_critical_pct"];
	        this.jitter_warning_ms = source["jitter_warning_ms"];
	        this.wifi_signal_warning = source["wifi_signal_warning"];
	        this.wifi_signal_critical = source["wifi_signal_critical"];
	        this.notifications_enabled = source["notifications_enabled"];
	        this.notification_cooldown_sec = source["notification_cooldown_sec"];
	    }
	}
	export class DiagnosisResponse {
	    id: number;
	    timestamp: string;
	    category: string;
	    severity: string;
	    title: string;
	    description: string;
	    evidence: storage.Evidence[];
	    resolved: boolean;
	    resolved_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosisResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.evidence = this.convertValues(source["evidence"], storage.Evidence);
	        this.resolved = source["resolved"];
	        this.resolved_at = source["resolved_at"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProbeResultResponse {
	    timestamp: string;
	    probe_type: string;
	    target: string;
	    success: boolean;
	    latency_ms: number;
	    jitter_ms: number;
	    packet_loss: number;
	
	    static createFrom(source: any = {}) {
	        return new ProbeResultResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.probe_type = source["probe_type"];
	        this.target = source["target"];
	        this.success = source["success"];
	        this.latency_ms = source["latency_ms"];
	        this.jitter_ms = source["jitter_ms"];
	        this.packet_loss = source["packet_loss"];
	    }
	}
	export class ProjectReportData {
	    totalProbes: number;
	    totalDiagnoses: number;
	    totalSpeedTests: number;
	    totalWifiSnapshots: number;
	    totalNetworkEvents: number;
	    uptimeHour: number;
	    uptime24h: number;
	    uptime7d: number;
	    avgLatency: number;
	    avgGatewayLatency: number;
	    avgDnsLatency: number;
	    avgDownload: number;
	    avgUpload: number;
	    avgSignal: number;
	    bandHops: number;
	    currentSSID: string;
	    currentBand: string;
	    currentChannel: number;
	    monitoringSince: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectReportData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalProbes = source["totalProbes"];
	        this.totalDiagnoses = source["totalDiagnoses"];
	        this.totalSpeedTests = source["totalSpeedTests"];
	        this.totalWifiSnapshots = source["totalWifiSnapshots"];
	        this.totalNetworkEvents = source["totalNetworkEvents"];
	        this.uptimeHour = source["uptimeHour"];
	        this.uptime24h = source["uptime24h"];
	        this.uptime7d = source["uptime7d"];
	        this.avgLatency = source["avgLatency"];
	        this.avgGatewayLatency = source["avgGatewayLatency"];
	        this.avgDnsLatency = source["avgDnsLatency"];
	        this.avgDownload = source["avgDownload"];
	        this.avgUpload = source["avgUpload"];
	        this.avgSignal = source["avgSignal"];
	        this.bandHops = source["bandHops"];
	        this.currentSSID = source["currentSSID"];
	        this.currentBand = source["currentBand"];
	        this.currentChannel = source["currentChannel"];
	        this.monitoringSince = source["monitoringSince"];
	    }
	}
	export class SpeedTestResponse {
	    timestamp: string;
	    download_mbps: number;
	    upload_mbps: number;
	    latency_ms: number;
	    jitter_ms: number;
	    server: string;
	    triggered_by: string;
	
	    static createFrom(source: any = {}) {
	        return new SpeedTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.download_mbps = source["download_mbps"];
	        this.upload_mbps = source["upload_mbps"];
	        this.latency_ms = source["latency_ms"];
	        this.jitter_ms = source["jitter_ms"];
	        this.server = source["server"];
	        this.triggered_by = source["triggered_by"];
	    }
	}
	export class StatusResponse {
	    status: string;
	    category: string;
	    title: string;
	    description: string;
	    evidence: storage.Evidence[];
	    confidence: number;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new StatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.category = source["category"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.evidence = this.convertValues(source["evidence"], storage.Evidence);
	        this.confidence = source["confidence"];
	        this.timestamp = source["timestamp"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UptimeStats {
	    one_hour: number;
	    twenty_four_h: number;
	    seven_days: number;
	
	    static createFrom(source: any = {}) {
	        return new UptimeStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.one_hour = source["one_hour"];
	        this.twenty_four_h = source["twenty_four_h"];
	        this.seven_days = source["seven_days"];
	    }
	}
	export class WifiInfo {
	    interface: string;
	    ssid: string;
	    bssid: string;
	    frequency_mhz: number;
	    channel: number;
	    signal_dbm: number;
	    noise_dbm: number;
	    link_speed_mbps: number;
	    band: string;
	    signal_quality: string;
	
	    static createFrom(source: any = {}) {
	        return new WifiInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interface = source["interface"];
	        this.ssid = source["ssid"];
	        this.bssid = source["bssid"];
	        this.frequency_mhz = source["frequency_mhz"];
	        this.channel = source["channel"];
	        this.signal_dbm = source["signal_dbm"];
	        this.noise_dbm = source["noise_dbm"];
	        this.link_speed_mbps = source["link_speed_mbps"];
	        this.band = source["band"];
	        this.signal_quality = source["signal_quality"];
	    }
	}

}

export namespace scanner {
	
	export class Device {
	    ip: string;
	    mac: string;
	    vendor: string;
	    hostname: string;
	    interface: string;
	    // Go type: time
	    first_seen: any;
	    // Go type: time
	    last_seen: any;
	    is_gateway: boolean;
	    is_local: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.mac = source["mac"];
	        this.vendor = source["vendor"];
	        this.hostname = source["hostname"];
	        this.interface = source["interface"];
	        this.first_seen = this.convertValues(source["first_seen"], null);
	        this.last_seen = this.convertValues(source["last_seen"], null);
	        this.is_gateway = source["is_gateway"];
	        this.is_local = source["is_local"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace storage {
	
	export class Baseline {
	    id: number;
	    probe_type: string;
	    target: string;
	    // Go type: time
	    period_start: any;
	    // Go type: time
	    period_end: any;
	    p50_latency_ms: number;
	    p95_latency_ms: number;
	    avg_latency_ms: number;
	    packet_loss_rate: number;
	    sample_count: number;
	
	    static createFrom(source: any = {}) {
	        return new Baseline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.probe_type = source["probe_type"];
	        this.target = source["target"];
	        this.period_start = this.convertValues(source["period_start"], null);
	        this.period_end = this.convertValues(source["period_end"], null);
	        this.p50_latency_ms = source["p50_latency_ms"];
	        this.p95_latency_ms = source["p95_latency_ms"];
	        this.avg_latency_ms = source["avg_latency_ms"];
	        this.packet_loss_rate = source["packet_loss_rate"];
	        this.sample_count = source["sample_count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DiagnosisPeriod {
	    // Go type: time
	    start: any;
	    // Go type: time
	    end: any;
	    category: string;
	    severity: string;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosisPeriod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = this.convertValues(source["start"], null);
	        this.end = this.convertValues(source["end"], null);
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Evidence {
	    type: string;
	    description: string;
	    value: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Evidence(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.description = source["description"];
	        this.value = source["value"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HeatmapCell {
	    hour: number;
	    probe_type: string;
	    success_rate: number;
	    avg_latency: number;
	    sample_count: number;
	
	    static createFrom(source: any = {}) {
	        return new HeatmapCell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
	        this.probe_type = source["probe_type"];
	        this.success_rate = source["success_rate"];
	        this.avg_latency = source["avg_latency"];
	        this.sample_count = source["sample_count"];
	    }
	}
	export class NetworkEvent {
	    id: number;
	    // Go type: time
	    timestamp: any;
	    reason: string;
	    prev_ssid: string;
	    prev_type: string;
	    prev_interface: string;
	    prev_gateway: string;
	    curr_ssid: string;
	    curr_type: string;
	    curr_interface: string;
	    curr_gateway: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.reason = source["reason"];
	        this.prev_ssid = source["prev_ssid"];
	        this.prev_type = source["prev_type"];
	        this.prev_interface = source["prev_interface"];
	        this.prev_gateway = source["prev_gateway"];
	        this.curr_ssid = source["curr_ssid"];
	        this.curr_type = source["curr_type"];
	        this.curr_interface = source["curr_interface"];
	        this.curr_gateway = source["curr_gateway"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProbeLatencyPoint {
	    // Go type: time
	    timestamp: any;
	    gateway_ms: number;
	    external_ms: number;
	    dns_ms: number;
	    tcp_ms: number;
	    gateway_loss: number;
	    external_loss: number;
	
	    static createFrom(source: any = {}) {
	        return new ProbeLatencyPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.gateway_ms = source["gateway_ms"];
	        this.external_ms = source["external_ms"];
	        this.dns_ms = source["dns_ms"];
	        this.tcp_ms = source["tcp_ms"];
	        this.gateway_loss = source["gateway_loss"];
	        this.external_loss = source["external_loss"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TimeseriesPoint {
	    // Go type: time
	    timestamp: any;
	    value: number;
	    label?: string;
	
	    static createFrom(source: any = {}) {
	        return new TimeseriesPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.value = source["value"];
	        this.label = source["label"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WifiTimeseriesPoint {
	    // Go type: time
	    timestamp: any;
	    signal_dbm: number;
	    channel: number;
	    frequency_mhz: number;
	    link_speed_mbps: number;
	    band: string;
	
	    static createFrom(source: any = {}) {
	        return new WifiTimeseriesPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.signal_dbm = source["signal_dbm"];
	        this.channel = source["channel"];
	        this.frequency_mhz = source["frequency_mhz"];
	        this.link_speed_mbps = source["link_speed_mbps"];
	        this.band = source["band"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

