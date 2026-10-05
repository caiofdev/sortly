import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import App from './App';

const criteria = [
  { key: 'byDuration', enabled: false, locked: false },
  { key: 'byPages', enabled: false, locked: false },
  { key: 'byResolution', enabled: false, locked: false },
  { key: 'byDate', enabled: false, locked: false },
  { key: 'bySize', enabled: false, locked: false },
  { key: 'byExtension', enabled: true, locked: true }
];

// Cada estado criado é mais novo que o anterior, como no backend.
let lastVersion = 0;

const viewState = (overrides = {}) => ({
  version: ++lastVersion,
  sourceFolderPath: '',
  destinationFolderPath: '',
  hasUndo: false,
  busy: '',
  settings: { language: 'pt-BR', criteria },
  notifications: [],
  ...overrides
});

function mockBackend(initial, bindings = {}) {
  window.go = {
    app: { App: { GetState: vi.fn().mockResolvedValue(initial), ...bindings } }
  };
}

describe('App', () => {
  beforeEach(() => {
    window.runtime = {
      OnFileDrop: vi.fn(),
      OnFileDropOff: vi.fn(),
      EventsOnMultiple: vi.fn(() => vi.fn())
    };
  });

  it('só renderiza depois do estado do backend, em português por padrão', async () => {
    mockBackend(viewState());
    const { container } = render(<App />);

    expect(container).toBeEmptyDOMElement();
    expect(await screen.findByRole('heading', { name: 'Sortly' })).toBeInTheDocument();
    expect(screen.getByText('Arraste e solte uma pasta ou arquivo aqui')).toBeInTheDocument();
  });

  it('usa o idioma das preferências, sem passar pelo português', async () => {
    mockBackend(viewState({ settings: { language: 'en', criteria } }));
    render(<App />);

    expect(await screen.findByText('Drag and drop a folder or file here')).toBeInTheDocument();
    expect(screen.queryByText('Arraste e solte uma pasta ou arquivo aqui')).not.toBeInTheDocument();
  });

  it('mostra caminhos, carregamento e notificações traduzidas do estado', async () => {
    mockBackend(
      viewState({
        sourceFolderPath: 'C:\\origem',
        busy: 'organize',
        notifications: [
          {
            id: 1,
            kind: 'info',
            code: 'RECOVERED_LAST_ORGANIZATION',
            action: 'startup',
            at: '2026-03-05T12:00:00Z'
          }
        ]
      })
    );
    render(<App />);

    expect(await screen.findByText('C:\\origem')).toBeInTheDocument();
    expect(screen.getByText('Organizando arquivos...')).toBeInTheDocument();
    expect(
      screen.getByText('Última organização recuperada. Você pode desfazer essa alteração.')
    ).toBeInTheDocument();
  });

  it('ações chamam os bindings e mostram o estado devolvido', async () => {
    const initial = viewState({ sourceFolderPath: 'C:\\origem' });
    const organized = viewState({ sourceFolderPath: 'C:\\origem', hasUndo: true });
    mockBackend(initial, {
      Organize: vi.fn().mockResolvedValue(organized)
    });
    render(<App />);

    const organize = await screen.findByRole('button', { name: 'Organizar arquivos' });
    fireEvent.click(organize);

    await vi.waitFor(() => expect(window.go.app.App.Organize).toHaveBeenCalled());
    expect(await screen.findByRole('button', { name: 'Desfazer ultima separação' })).toBeEnabled();
  });
});
