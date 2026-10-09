import { useEffect, useRef } from 'react';
import { FolderIcon, UndoIcon } from './icons';
import { folderLabel } from '../i18n/folderLabel';
import { folderName } from './OrganizingPanel';

// As barras são relativas à maior pasta, como no protótipo (#79).
export const barWidth = (count, max) => `${max > 0 ? Math.round((count / max) * 100) : 0}%`;

function FolderBar({ name, count, max }) {
  return (
    <div className="st-done__row">
      <span className="st-done__folder">{name}</span>
      <div className="st-done__track">
        <div className="st-done__bar" style={{ width: barWidth(count, max) }} />
      </div>
      <span className="st-done__count">{count}</span>
    </div>
  );
}

function DonePanel({
  labels,
  result,
  hasUndo,
  isLoading,
  loadingAction,
  onOpenDestination,
  onUndo,
  onStartOver
}) {
  const max = Math.max(result.otherFiles, ...result.folders.map((f) => f.count));
  const restoring = loadingAction === 'restore';
  const head = useRef(null);

  // O Cancelar, que tinha o foco, some quando a organização acaba; o foco vai
  // para o resumo, e o leitor de tela anuncia o resultado (#79).
  useEffect(() => {
    head.current.focus();
  }, []);

  return (
    <div className="st-done">
      <div className="st-done__head" ref={head} tabIndex={-1}>
        <p className="st-done__number">{result.movedFiles}</p>
        <div className="st-done__text">
          <p className="st-done__title">
            {result.movedFiles === 1 ? labels.doneTitleOne : labels.doneTitle}
          </p>
          <p className="st-done__sub">
            {labels.doneIn} {folderName(result.destinationFolderPath)}
          </p>
          {result.failedFiles > 0 && (
            <p className="st-done__sub st-done__sub--err">
              {result.failedFiles}{' '}
              {result.failedFiles === 1 ? labels.doneFailedOne : labels.doneFailed}
            </p>
          )}
        </div>
      </div>

      <div className="st-done__folders">
        {result.folders.map((f) => (
          <FolderBar
            key={f.name}
            name={folderLabel(f.name, result.categoryFolders, labels)}
            count={f.count}
            max={max}
          />
        ))}
        {result.otherFiles > 0 && (
          <FolderBar name={labels.previewOthers} count={result.otherFiles} max={max} />
        )}
      </div>

      <div className="st-done__actions">
        <button
          type="button"
          className="st-btn st-btn--secondary st-btn--lg"
          onClick={onOpenDestination}
          disabled={isLoading}
        >
          <FolderIcon />
          {labels.openDestination}
        </button>
        <button
          type="button"
          className="st-btn st-btn--danger st-btn--lg"
          onClick={onUndo}
          disabled={!hasUndo || isLoading}
          aria-busy={restoring}
        >
          <UndoIcon />
          {restoring ? labels.restoring : labels.undo}
        </button>
        <button
          type="button"
          className="st-btn st-btn--primary st-btn--lg"
          onClick={onStartOver}
          disabled={isLoading}
        >
          {labels.startOver}
        </button>
      </div>
    </div>
  );
}

export default DonePanel;
