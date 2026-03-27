import { useState } from 'react';

function DragDropPanel({ isLoading, labels, onResolveDroppedPath, onSelectSourceFolder }) {
  const [isDragging, setIsDragging] = useState(false);

  const handleDragOver = (event) => {
    event.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (event) => {
    event.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (event) => {
    event.preventDefault();
    setIsDragging(false);

    const droppedItem = event.dataTransfer?.files?.[0];
    const droppedPath = droppedItem?.path;

    if (!droppedPath) {
      return;
    }

    onResolveDroppedPath(droppedPath);
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className={`rounded-2xl border-2 border-dashed px-6 py-10 text-center transition-colors ${
        isDragging
          ? 'border-[#3B82F6] bg-[#3B82F6]/12'
          : 'border-white/20 bg-[#0F172A]/60'
      } ${isLoading ? 'opacity-60' : ''}`}
    >
      <p className="text-lg font-semibold text-[#F8FAFC]">{labels.dropTitle}</p>
      <p className="mt-2 text-sm text-[#94A3B8]">{labels.dropDescription}</p>

      <p className="mt-2 text-sm text-[#94A3B8]">
        {labels.dropSelectHintPrefix}{' '}
        <button
          type="button"
          onClick={onSelectSourceFolder}
          className="font-semibold text-[#3B82F6] underline underline-offset-2 transition-colors hover:text-[#60a5fa]"
        >
          {labels.dropSelectHintAction}
        </button>
      </p>
    </div>
  );
}

export default DragDropPanel;
