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

function HistoryPage({ labels, history, language }) {
  return (
    <section className="st-card st-history" aria-label={labels.titleHistory}>
      <div className="st-history__table">
        <div className="st-history__row st-history__row--head">
          <span>{labels.historyWhen}</span>
          <span>{labels.sourceLabel}</span>
          <span className="st-history__files">{labels.historyFiles}</span>
          <span>{labels.historyStatus}</span>
        </div>
        {history.length === 0 && <p className="st-history__empty">{labels.historyEmpty}</p>}
        {history.map((entry) => {
          const status = STATUS[entry.status] ?? STATUS.done;
          return (
            <div key={entry.at} className="st-history__row">
              <span className="st-history__when">{formatWhen(entry.at, language)}</span>
              <span className="st-history__path" title={entry.sourceFolderPath}>
                {entry.sourceFolderPath}
              </span>
              <span className="st-history__files st-history__count">{entry.movedFiles}</span>
              <span>
                <span className={status.className}>{labels[status.labelKey]}</span>
              </span>
            </div>
          );
        })}
      </div>
    </section>
  );
}

export default HistoryPage;
