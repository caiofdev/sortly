import { describe, expect, it } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import HistoryPage, { formatWhen } from './HistoryPage';

const pt = organizerCopy['pt-BR'];
const at = new Date(2026, 2, 5, 18, 40).toISOString();

const entry = (overrides = {}) => ({
  at,
  sourceFolderPath: 'C:\\Users\\caio\\Downloads',
  destinationFolderPath: '',
  movedFiles: 42,
  status: 'done',
  ...overrides
});

function renderHistory(history, language = 'pt-BR') {
  return render(<HistoryPage labels={pt} history={history} language={language} />);
}

describe('formatWhen', () => {
  it.each([
    ['pt-BR', /5 de mar\. de 2026.*18:40/],
    ['en', /Mar 5, 2026.*6:40/]
  ])('%s: data e hora no idioma escolhido', (language, expected) => {
    expect(formatWhen(at, language)).toMatch(expected);
  });
});

describe('HistoryPage', () => {
  it('sem histórico: cabeçalho e o texto de vazio', () => {
    renderHistory([]);
    expect(screen.getByText(pt.historyWhen)).toBeInTheDocument();
    expect(screen.getByText(pt.historyEmpty)).toBeInTheDocument();
  });

  it('uma linha por organização, com a pasta inteira no title', () => {
    renderHistory([entry(), entry({ at: new Date(2026, 2, 4).toISOString(), movedFiles: 7 })]);
    expect(screen.queryByText(pt.historyEmpty)).not.toBeInTheDocument();
    const table = screen.getByRole('table', { name: pt.titleHistory });
    expect(within(table).getAllByRole('row')).toHaveLength(3);
    expect(
      within(table)
        .getAllByRole('columnheader')
        .map((c) => c.textContent)
    ).toEqual([pt.historyWhen, pt.sourceLabel, pt.historyFiles, pt.historyStatus]);
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getAllByTitle('C:\\Users\\caio\\Downloads')).toHaveLength(2);
  });

  it.each([
    ['done', pt.historyDone, 'st-status--ok'],
    ['undone', pt.historyUndone, 'st-status--undo'],
    ['canceled', pt.historyCanceled, 'st-status--undo'],
    ['desconhecido', pt.historyDone, 'st-status--ok']
  ])('status %s: selo "%s" (%s)', (status, label, className) => {
    renderHistory([entry({ status })]);
    expect(screen.getByText(label)).toHaveClass('st-status', className);
  });
});
