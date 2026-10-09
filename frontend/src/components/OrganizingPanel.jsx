import { useEffect, useRef } from 'react';
import FileFlow from './FileFlow';
import ProgressBlocks from './ProgressBlocks';

// Só o nome da pasta, como no protótipo; o caminho inteiro está na tela anterior (#78).
export const folderName = (path) => path.split(/[\\/]/).filter(Boolean).pop() ?? path;

function OrganizingPanel({ labels, progress, sourceFolderPath, destinationFolderPath, onCancel }) {
  const { done, total, file, folder } = progress;
  const cancel = useRef(null);

  // O botão Organizar, que tinha o foco, some quando o painel aparece; sem isto o
  // foco cairia no <body> e o teclado perderia o Cancelar (#78).
  useEffect(() => {
    cancel.current.focus();
  }, []);

  return (
    <div className="st-running">
      <div className="st-running__head">
        <p className="st-running__title">{labels.organizing}</p>
        <p className="st-running__count">
          {total > 0 ? `${done} / ${total}` : labels.progressPreparing}
        </p>
      </div>

      <FileFlow
        done={done}
        total={total}
        sourceName={folderName(sourceFolderPath)}
        destinationName={folderName(destinationFolderPath || sourceFolderPath)}
      />

      <ProgressBlocks label={labels.progressLabel} done={done} total={total} />

      <p className="st-running__now">
        {file && (
          <>
            <span className="st-running__muted">{labels.moving}</span>{' '}
            <span className="st-running__file">{file}</span>{' '}
            <span className="st-running__muted">→</span>{' '}
            <span className="st-running__folder">{folder || labels.previewRoot}</span>
          </>
        )}
      </p>

      <div className="st-actions">
        <button ref={cancel} type="button" className="st-btn st-btn--ghost" onClick={onCancel}>
          {labels.cancel}
        </button>
      </div>
    </div>
  );
}

export default OrganizingPanel;
