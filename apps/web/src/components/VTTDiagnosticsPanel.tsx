import { useTranslation } from "react-i18next";
import {
  downloadTelemetryReport,
  type VTTTelemetryReport,
} from "../telemetry";

interface Props {
  report: VTTTelemetryReport;
}

function Metric({ label, value, testId }: { label: string; value: string | number; testId?: string }) {
  return (
    <div className="diagnostic-metric">
      <span>{label}</span>
      <strong data-testid={testId}>{value}</strong>
    </div>
  );
}
export function VTTDiagnosticsPanel({ report }: Props) {
  const { t } = useTranslation();
  return (
    <aside className="diagnostics-panel" data-testid="vtt-diagnostics-panel">
      <header>
        <div>
          <small>{t("diagnosticsLocal")}</small>
          <h2>{t("diagnostics")}</h2>
        </div>
        <i className={report.connection.status === "open" ? "online" : ""} />
      </header>

      <section>
        <h3>{t("diagnosticsFrames")}</h3>
        <div className="diagnostic-grid">
          <Metric label="FPS" value={report.frames.fps} />
          <Metric label="AVG" value={`${report.frames.averageFrameMs} ms`} />
          <Metric label="P95" value={`${report.frames.p95FrameMs} ms`} />
          <Metric label={t("diagnosticsLongFrames")} value={report.frames.longFrames} />
          <Metric label={t("diagnosticsSamples")} value={report.frames.sampleCount} testId="telemetry-samples" />
        </div>
      </section>

      <section>
        <h3>Renderer</h3>
        <div className="diagnostic-grid">
          <Metric label="Draw calls" value={report.renderer.drawCalls} />
          <Metric label={t("diagnosticsTriangles")} value={report.renderer.triangles} />
          <Metric label={t("diagnosticsGeometries")} value={report.renderer.geometries} />
          <Metric label={t("diagnosticsTextures")} value={report.renderer.textures} />
        </div>
      </section>

      <section>
        <h3>{t("diagnosticsScene")}</h3>
        <div className="diagnostic-grid">
          <Metric label="Grid" value={`${report.scene.width}×${report.scene.height}`} />
          <Metric label="Tiles" value={report.scene.tiles} />
          <Metric label={t("diagnosticsWalls")} value={report.scene.walls} />
          <Metric label="Props" value={report.scene.props} />
          <Metric label="Tokens" value={report.scene.tokens} />
          <Metric label="Fog" value={report.scene.fogCells} />
        </div>
      </section>

      <section>
        <h3>{t("diagnosticsConnection")}</h3>
        <div className="diagnostic-grid">
          <Metric label="Status" value={report.connection.status} />
          <Metric label={t("diagnosticsRevision")} value={report.connection.lastRevision} />
          <Metric label={t("diagnosticsCommands")} value={report.connection.commandsSent} />
          <Metric label={t("diagnosticsEvents")} value={report.connection.eventsReceived} />
          <Metric label="Reconnects" value={report.connection.reconnectAttempts} />
          <Metric label={t("diagnosticsRejected")} value={report.connection.rejectedCommands} />
        </div>
      </section>

      <button className="secondary diagnostics-download" onClick={() => downloadTelemetryReport(report)}>
        JSON↓ <span>{t("diagnosticsDownload")}</span>
      </button>
    </aside>
  );
}
