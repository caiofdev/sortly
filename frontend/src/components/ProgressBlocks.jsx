export const BLOCKS = 32;

// Só blocos cheios: 31/32 do caminho ainda mostra 31 blocos, nunca a barra
// completa antes do fim (design system, Progress; #78).
export function filledBlocks(done, total) {
  if (total <= 0) return 0;
  return Math.floor((done / total) * BLOCKS);
}

function ProgressBlocks({ label, done, total }) {
  const filled = filledBlocks(done, total);
  return (
    <div
      className="st-blocks"
      role="progressbar"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={total}
      aria-valuenow={done}
    >
      {Array.from({ length: BLOCKS }, (_, i) => (
        <span key={i} className={i < filled ? 'st-blocks__b st-blocks__b--on' : 'st-blocks__b'} />
      ))}
    </div>
  );
}

export default ProgressBlocks;
