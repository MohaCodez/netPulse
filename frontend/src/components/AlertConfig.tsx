import { useState, useEffect } from 'react';
import './AlertConfig.css';

interface AlertRules {
  latency_warning_ms: number;
  latency_critical_ms: number;
  loss_warning_pct: number;
  loss_critical_pct: number;
  jitter_warning_ms: number;
  wifi_signal_warning: number;
  wifi_signal_critical: number;
  notifications_enabled: boolean;
  notification_cooldown_sec: number;
}

const defaultRules: AlertRules = {
  latency_warning_ms: 150,
  latency_critical_ms: 500,
  loss_warning_pct: 10,
  loss_critical_pct: 50,
  jitter_warning_ms: 50,
  wifi_signal_warning: -70,
  wifi_signal_critical: -80,
  notifications_enabled: true,
  notification_cooldown_sec: 60,
};

export function AlertConfig() {
  const [rules, setRules] = useState<AlertRules>(defaultRules);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (window.go?.main?.App?.GetAlertRules) {
      window.go.main.App.GetAlertRules().then((r: AlertRules | null) => {
        if (r) setRules(r);
      });
    }
  }, []);

  const handleSave = async () => {
    if (window.go?.main?.App?.SetAlertRules) {
      await window.go.main.App.SetAlertRules(rules);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    }
  };

  const handleReset = () => {
    setRules(defaultRules);
  };

  const updateRule = (key: keyof AlertRules, value: number | boolean) => {
    setRules((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <div className="alert-config">
      <div className="alert-header">
        <h3>⚙️ Alert Rules</h3>
        <div className="alert-actions">
          <button className="alert-btn btn-reset" onClick={handleReset}>Reset</button>
          <button className="alert-btn btn-save" onClick={handleSave}>
            {saved ? '✓ Saved' : 'Save'}
          </button>
        </div>
      </div>

      <div className="rule-groups">
        <div className="rule-group">
          <h4>Latency</h4>
          <div className="rule-row">
            <label>Warning</label>
            <div className="rule-input">
              <input type="number" value={rules.latency_warning_ms}
                onChange={(e) => updateRule('latency_warning_ms', Number(e.target.value))} />
              <span>ms</span>
            </div>
          </div>
          <div className="rule-row">
            <label>Critical</label>
            <div className="rule-input">
              <input type="number" value={rules.latency_critical_ms}
                onChange={(e) => updateRule('latency_critical_ms', Number(e.target.value))} />
              <span>ms</span>
            </div>
          </div>
        </div>

        <div className="rule-group">
          <h4>Packet Loss</h4>
          <div className="rule-row">
            <label>Warning</label>
            <div className="rule-input">
              <input type="number" value={rules.loss_warning_pct}
                onChange={(e) => updateRule('loss_warning_pct', Number(e.target.value))} />
              <span>%</span>
            </div>
          </div>
          <div className="rule-row">
            <label>Critical</label>
            <div className="rule-input">
              <input type="number" value={rules.loss_critical_pct}
                onChange={(e) => updateRule('loss_critical_pct', Number(e.target.value))} />
              <span>%</span>
            </div>
          </div>
        </div>

        <div className="rule-group">
          <h4>Jitter</h4>
          <div className="rule-row">
            <label>Warning</label>
            <div className="rule-input">
              <input type="number" value={rules.jitter_warning_ms}
                onChange={(e) => updateRule('jitter_warning_ms', Number(e.target.value))} />
              <span>ms</span>
            </div>
          </div>
        </div>

        <div className="rule-group">
          <h4>Wi-Fi Signal</h4>
          <div className="rule-row">
            <label>Warning</label>
            <div className="rule-input">
              <input type="number" value={rules.wifi_signal_warning}
                onChange={(e) => updateRule('wifi_signal_warning', Number(e.target.value))} />
              <span>dBm</span>
            </div>
          </div>
          <div className="rule-row">
            <label>Critical</label>
            <div className="rule-input">
              <input type="number" value={rules.wifi_signal_critical}
                onChange={(e) => updateRule('wifi_signal_critical', Number(e.target.value))} />
              <span>dBm</span>
            </div>
          </div>
        </div>

        <div className="rule-group">
          <h4>Notifications</h4>
          <div className="rule-row">
            <label>Enabled</label>
            <input type="checkbox" checked={rules.notifications_enabled}
              onChange={(e) => updateRule('notifications_enabled', e.target.checked)} />
          </div>
          <div className="rule-row">
            <label>Cooldown</label>
            <div className="rule-input">
              <input type="number" value={rules.notification_cooldown_sec}
                onChange={(e) => updateRule('notification_cooldown_sec', Number(e.target.value))} />
              <span>sec</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
