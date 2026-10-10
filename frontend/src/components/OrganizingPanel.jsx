import { useEffect, useRef } from 'react';
import FileFlow from './FileFlow';
import ProgressBlocks from './ProgressBlocks';
import { folderPathLabel } from '../i18n/folderLabel';

// Só o nome da pasta, como no protótipo; o caminho inteiro está na tela anterior (#78).
export const folderName = (path) => path.split(/[\\/]/).filter(Boolean).pop() ?? path;

// Desfazer usa a mesma tela no sentido inverso: os arquivos saem do destino e
// voltam para a origem (#96).
const MODES = {
  organize: { title: 'organizing', moving: 'moving', progress: 'progressLabel', reverse: false },
  restore: {
    title: 'restoring',
    moving: 'restoreMoving',
    progress: 'restoreProgressLabel',
    reverse: true
  }
};

function OrganizingPanel({
  labels,
  mode = 'organize',
  progress,
  sourceFolderPath,
  destinationFolderPath,
  onCancel
}) {
  const { done, total, file, folder, categoryFolders } = progress;
  const text = MODES[mode];
  const source = folderName(sourceFolderPath);
  const destination = folderName(destinationFolderPath || sourceFolderPath);
  const [from, to] = text.reverse ? [destination, source] : [source, destination];
  const cancel = useRef(null);

  // O botão Organizar, que tinha o foco, some quando o painel aparece; sem isto o
  // foco cairia no <body> e o teclado perderia o Cancelar (#78).
  useEffect(() => {
    cancel.current.focus();
  }, []);

  return (
    <div className="st-running">
      <div className="st-running__head">
        <p className="st-running__title">{labels[text.title]}</p>
        <p className="st-running__count">
          {total > 0 ? `${done} / ${total}` : labels.progressPreparing}
        </p>
      </div>

      <FileFlow done={done} total={total} sourceName={from} destinationName={to} />

      <ProgressBlocks label={labels[text.progress]} done={done} total={total} />

      <p className="st-running__now">
        {file && (
          <>
            <span className="st-running__muted">{labels[text.moving]}</span>{' '}
            <span className="st-running__file">{file}</span>{' '}
            <span className="st-running__muted">→</span>{' '}
            <span className="st-running__folder">
              {folderPathLabel(folder, categoryFolders, labels)}
            </span>
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
