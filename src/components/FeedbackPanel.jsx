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
    feedback.type === 'organize'
      ? 'border-[#22C55E]/45 bg-[#22C55E]/12 text-[#b8ffcf]'
      : feedback.type === 'restore'
        ? 'border-rose-500/45 bg-rose-500/12 text-rose-200'
        : feedback.type === 'error'
          ? 'border-rose-500/45 bg-rose-500/12 text-rose-200'
          : 'border-[#3B82F6]/45 bg-[#3B82F6]/12 text-[#bfdbfe]';

  return (
    <div
      className={`fixed bottom-5 right-5 z-50 w-[min(92vw,420px)] rounded-2xl border px-4 py-4 text-sm shadow-xl backdrop-blur-sm transition-all duration-300 ${toneClass} ${
        isVisible ? 'translate-y-0 opacity-100' : 'translate-y-2 opacity-0'
      }`}
    >
      <p>{feedback.message}</p>
    </div>
  );
}

export default FeedbackPanel;
