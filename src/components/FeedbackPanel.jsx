import { useEffect, useState } from 'react';

function FeedbackPanel({ feedback, onClearFeedback }) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    if (!feedback) {
      setIsVisible(false);
      return;
    }

    setIsVisible(true);

    const startHideId = window.setTimeout(() => {
      setIsVisible(false);
    }, 2600);

    const clearId = window.setTimeout(() => {
      onClearFeedback();
    }, 3000);

    return () => {
      window.clearTimeout(startHideId);
      window.clearTimeout(clearId);
    };
  }, [feedback, onClearFeedback]);

  if (!feedback) {
    return null;
  }

  const toneClass =
    feedback.type === 'success'
      ? 'border-[#22C55E]/45 bg-[#22C55E]/12 text-[#b8ffcf]'
      : 'border-rose-500/45 bg-rose-500/12 text-rose-200';

  return (
    <div
      className={`rounded-2xl border px-4 py-4 text-sm shadow-sm transition-all duration-300 ${toneClass} ${
        isVisible ? 'translate-y-0 opacity-100' : '-translate-y-1 opacity-0'
      }`}
    >
      <p>{feedback.message}</p>
    </div>
  );
}

export default FeedbackPanel;
