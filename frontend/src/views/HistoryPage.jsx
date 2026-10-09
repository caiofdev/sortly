import { toLocale } from '../i18n/language';

// O status vem do backend; aqui só o texto e a cor do selo (#80).
const STATUS = {
  done: { labelKey: 'historyDone', className: 'st-status st-status--ok' },
  undone: { labelKey: 'historyUndone', className: 'st-status st-status--undo' },
  canceled: { labelKey: 'historyCanceled', className: 'st-status st-status--undo' }
};

export function formatWhen(at, language) {
  return new Date(at).toLocaleString(toLocale(language), {
    dateStyle: 'medium',
    timeStyle: 'short'
  });
}

// A grade do protótipo é feita de divs; os papéis ARIA deixam o leitor de tela
// navegá-la como tabela, com o nome da coluna em cada célula (#80).
function HistoryPage({ labels, history, language }) {
  return (
    <section className="st-card st-history">
      <div className="st-history__table" role="table" aria-label={labels.titleHistory}>
        <div className="st-history__row st-history__row--head" role="row">
          <span role="columnheader">{labels.historyWhen}</span>
          <span role="columnheader">{labels.sourceLabel}</span>
          <span role="columnheader" className="st-history__files">
            {labels.historyFiles}
          </span>
          <span role="columnheader">{labels.historyStatus}</span>
        </div>
        {history.map((entry, index) => {
          const status = STATUS[entry.status] ?? STATUS.done;
          return (
            <div key={`${entry.at}-${index}`} className="st-history__row" role="row">
              <span role="cell" className="st-history__when">
                {formatWhen(entry.at, language)}
              </span>
              <span role="cell" className="st-history__path" title={entry.sourceFolderPath}>
                {entry.sourceFolderPath}
              </span>
              <span role="cell" className="st-history__files st-history__count">
                {entry.movedFiles}
              </span>
              <span role="cell">
                <span className={status.className}>{labels[status.labelKey]}</span>
              </span>
            </div>
          );
        })}
      </div>
      {history.length === 0 && <p className="st-history__empty">{labels.historyEmpty}</p>}
    </section>
  );
}

export default HistoryPage;
