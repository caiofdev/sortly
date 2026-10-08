import { useRef, useState } from 'react';
import { DropFolderIcon } from './icons';

// Marca o painel como área de drop do Wails; a propriedade é herdada pelos filhos (#12).
const DROP_TARGET_STYLE = { '--wails-drop-target': 'drop' };

function DragDropPanel({ isLoading, labels, onSelectSourceFolder }) {
  const [isDragging, setIsDragging] = useState(false);
  // Entrar num elemento interno dispara dragleave no painel; contar entradas e
  // saídas evita o destaque piscar. O relatedTarget resolveria no Chromium, mas
  // o WebKit (macOS e Linux) costuma entregá-lo vazio (#60).
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
    // OnFileDrop assinado no useViewState. Aqui só desliga o destaque (#12).
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
      className={isDragging ? 'st-drop st-drop--active' : 'st-drop'}
      aria-busy={isLoading}
    >
      <DropFolderIcon />
      <p className="st-drop__title">{isDragging ? labels.dropActive : labels.dropTitle}</p>
      <p className="st-drop__hint">{labels.dropDescription}</p>
      <p className="st-drop__hint">
        {labels.dropSelectHintPrefix}{' '}
        <button
          type="button"
          className="st-link"
          onClick={onSelectSourceFolder}
          disabled={isLoading}
        >
          {labels.dropSelectHintAction}
        </button>
      </p>
    </div>
  );
}

export default DragDropPanel;
