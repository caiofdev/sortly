function FeedbackPanel({ feedback, emptyMessage }) {
  const toneClass = feedback
    ? feedback.type === 'success'
      ? 'border-[#22C55E]/45 bg-[#22C55E]/12 text-[#b8ffcf]'
      : 'border-rose-500/45 bg-rose-500/12 text-rose-200'
    : 'border-white/10 bg-[#0F172A]/65 text-[#94A3B8]';

  return (
    <div className={`rounded-2xl border px-4 py-4 text-sm shadow-sm ${toneClass}`}>
      {feedback ? feedback.message : emptyMessage}
    </div>
  );
}

export default FeedbackPanel;
