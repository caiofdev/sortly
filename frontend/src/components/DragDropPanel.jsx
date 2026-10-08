import { useRef, useState } from 'react';

// Marca o painel como área de drop do Wails (a propriedade é herdada pelos filhos).
const DROP_TARGET_STYLE = { '--wails-drop-target': 'drop' };

function DragDropPanel({ isLoading, labels, onSelectSourceFolder }) {
  const [isDragging, setIsDragging] = useState(false);
  // Entrar num elemento interno dispara dragleave no painel; contar entradas e
  // saídas evita o destaque piscar. O relatedTarget resolveria no Chromium, mas
  // o WebKit (macOS e Linux) costuma entregá-lo vazio.
  const depth = useRef(0);

  const handleDragEnter = (event) => {
    event.preventDefault();
    depth.current += 1;
    setIsDragging(true);
  };

  const handleDragOver = (event) => {
    event.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (event) => {
    event.preventDefault();
    depth.current = Math.max(0, depth.current - 1);
    if (depth.current === 0) setIsDragging(false);
  };

  const handleDrop = (event) => {
    event.preventDefault();
    // O caminho não vem no evento do navegador: o Wails o entrega pelo
    // OnFileDrop assinado no useViewState. Aqui só desliga o destaque.
    depth.current = 0;
    setIsDragging(false);
  };

  return (
    <div
      onDragEnter={handleDragEnter}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      style={DROP_TARGET_STYLE}
      className={`rounded-2xl border-2 border-dashed px-6 py-10 text-center transition-colors ${
        isDragging ? 'border-[#3B82F6] bg-[#3B82F6]/12' : 'border-white/20 bg-[#0F172A]/60'
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
