// Pixel art do design system (FileFlow): pasta 26×20 e arquivo 5×6, escalados 4×.
// Os arquivos "espiando" na boca de cada pasta acompanham o progresso real: somem
// da origem e aparecem no destino (#78).
const PEEKS = [
  {
    left: 22,
    bottom: 66,
    fill: 'var(--ink-200)',
    corner: 'var(--bg-000)',
    visible: (f) => f < 0.5
  },
  {
    left: 54,
    bottom: 70,
    fill: 'var(--ink-100)',
    corner: 'var(--bg-000)',
    visible: (f) => f < 0.86
  },
  {
    left: 440,
    bottom: 66,
    fill: 'var(--ink-100)',
    corner: 'var(--brand)',
    visible: (f) => f > 0.125
  },
  {
    left: 474,
    bottom: 70,
    fill: 'var(--ink-200)',
    corner: 'var(--brand)',
    visible: (f) => f > 0.625
  }
];

// Seis arquivos no mesmo arco, defasados 0,4 s; um deles é amarelo (#78).
const FLYING = [
  { delay: '0s', fill: 'var(--ink-100)', corner: 'var(--brand)', line: 'var(--bg-300)', short: 2 },
  {
    delay: '-0.4s',
    fill: 'var(--ink-100)',
    corner: 'var(--brand)',
    line: 'var(--bg-300)',
    short: 3
  },
  {
    delay: '-0.8s',
    fill: 'var(--brand)',
    corner: 'var(--brand-press)',
    line: 'var(--on-brand)',
    short: 2
  },
  {
    delay: '-1.2s',
    fill: 'var(--ink-100)',
    corner: 'var(--brand)',
    line: 'var(--bg-300)',
    short: 3
  },
  {
    delay: '-1.6s',
    fill: 'var(--ink-200)',
    corner: 'var(--brand)',
    line: 'var(--bg-300)',
    short: 2
  },
  {
    delay: '-2.0s',
    fill: 'var(--ink-100)',
    corner: 'var(--brand)',
    line: 'var(--bg-300)',
    short: 3
  }
];

const fraction = (done, total) => (total > 0 ? done / total : 0);

function FileFlow({ done, total, sourceName, destinationName }) {
  const f = fraction(done, total);
  return (
    <div className="st-flow-wrap">
      <div className="st-flow" aria-hidden="true">
        <svg className="st-flow__folder" viewBox="0 0 26 20" style={{ left: 0 }}>
          <rect x="0" y="0" width="11" height="4" style={{ fill: 'var(--line-200)' }} />
          <rect x="0" y="3" width="26" height="17" style={{ fill: 'var(--line-200)' }} />
          <rect x="1" y="4" width="24" height="15" style={{ fill: 'var(--bg-300)' }} />
          <rect x="1" y="4" width="24" height="2" style={{ fill: 'var(--line-200)' }} />
        </svg>

        {PEEKS.filter((peek) => peek.visible(f)).map((peek) => (
          <svg
            key={peek.left}
            className="st-flow__peek"
            viewBox="0 0 5 6"
            style={{ left: peek.left, bottom: peek.bottom }}
          >
            <rect width="5" height="6" style={{ fill: peek.fill }} />
            <rect x="4" width="1" height="1" style={{ fill: peek.corner }} />
          </svg>
        ))}

        {FLYING.map((file) => (
          <svg
            key={file.delay}
            className="st-flow__file"
            viewBox="0 0 5 6"
            style={{ animationDelay: file.delay }}
          >
            <rect width="5" height="6" style={{ fill: file.fill }} />
            <rect x="4" width="1" height="1" style={{ fill: file.corner }} />
            <rect x="1" y="2" width="3" height="1" style={{ fill: file.line }} />
            <rect x="1" y="4" width={file.short} height="1" style={{ fill: file.line }} />
          </svg>
        ))}

        <svg className="st-flow__folder" viewBox="0 0 26 20" style={{ left: 416 }}>
          <rect x="0" y="0" width="11" height="4" style={{ fill: 'var(--brand-press)' }} />
          <rect x="0" y="3" width="26" height="17" style={{ fill: 'var(--brand)' }} />
          <rect x="0" y="3" width="26" height="2" style={{ fill: 'var(--brand-press)' }} />
          <rect x="9" y="10" width="8" height="1" style={{ fill: 'var(--on-brand)' }} />
        </svg>
      </div>

      <div className="st-flow__labels">
        <span>
          {sourceName} · <span className="st-flow__left">{total - done}</span>
        </span>
        <span>
          {destinationName} · <span className="st-flow__done">{done}</span>
        </span>
      </div>
    </div>
  );
}

export default FileFlow;
