import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';
import App from '../src/App';

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
  unread: false,
  preview: { status: '', totalFiles: 0, folders: [], otherFiles: 0 },
  settings: { language: 'pt-BR', theme: 'dark', criteria },
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
    expect(await screen.findByRole('heading', { name: 'Organizar pasta' })).toBeInTheDocument();
    expect(screen.getByText('Arraste e solte uma pasta ou arquivo aqui')).toBeInTheDocument();
    expect(document.documentElement.lang).toBe('pt-BR');
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('o tema das preferências vai para o <html>, e Configurações o troca', async () => {
    const light = viewState({ settings: { language: 'pt-BR', theme: 'light', criteria } });
    const dark = viewState({ settings: { language: 'pt-BR', theme: 'dark', criteria } });
    mockBackend(light, { SetTheme: vi.fn().mockResolvedValue(dark) });
    render(<App />);

    await screen.findByRole('heading', { name: 'Organizar pasta' });
    expect(document.documentElement.dataset.theme).toBe('light');

    fireEvent.click(screen.getByRole('button', { name: 'Configurações' }));
    fireEvent.click(screen.getByRole('button', { name: 'Escuro' }));
    await vi.waitFor(() => expect(document.documentElement.dataset.theme).toBe('dark'));
    expect(window.go.app.App.SetTheme).toHaveBeenCalledWith('dark');
  });

  it('usa o idioma das preferências, sem passar pelo português', async () => {
    mockBackend(viewState({ settings: { language: 'en', criteria } }));
    render(<App />);

    expect(await screen.findByText('Drag and drop a folder or file here')).toBeInTheDocument();
    expect(screen.queryByText('Arraste e solte uma pasta ou arquivo aqui')).not.toBeInTheDocument();
    expect(document.documentElement.lang).toBe('en');
  });

  it('mostra caminhos, carregamento e notificações traduzidas do estado', async () => {
    const recovered = {
      id: 1,
      kind: 'info',
      code: 'RECOVERED_LAST_ORGANIZATION',
      action: 'startup',
      at: '2026-03-05T12:00:00Z'
    };
    mockBackend(
      viewState({
        sourceFolderPath: 'C:\\origem',
        busy: 'organize',
        unread: true,
        notifications: [recovered]
      })
    );
    window.go.app.App.MarkNotificationsRead = vi
      .fn()
      .mockResolvedValue(
        viewState({ sourceFolderPath: 'C:\\origem', busy: 'organize', notifications: [recovered] })
      );
    const { container } = render(<App />);

    expect(await screen.findByText('C:\\origem')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Organizando…' })).toBeDisabled();
    expect(container.querySelector('.st-toast')).toHaveTextContent(
      'Última organização recuperadaVocê pode desfazer essa alteração.'
    );
    const bell = screen.getByRole('button', { name: 'Notificações' });
    expect(bell).toHaveAccessibleDescription('Há avisos novos');

    fireEvent.click(bell);
    const panel = screen.getByRole('dialog', { name: 'Notificações' });
    expect(within(panel).getByText('Última organização recuperada')).toBeInTheDocument();
    await vi.waitFor(() =>
      expect(container.querySelector('.st-icon-btn__dot')).not.toBeInTheDocument()
    );
    expect(window.go.app.App.MarkNotificationsRead).toHaveBeenCalledTimes(1);
  });

  it('mostra a prévia da origem que vem do backend', async () => {
    mockBackend(
      viewState({
        sourceFolderPath: 'C:\\origem',
        preview: {
          status: 'ready',
          totalFiles: 3,
          folders: [{ name: 'pdf', count: 2 }],
          otherFiles: 1
        }
      })
    );
    render(<App />);

    expect(await screen.findByText('arquivos encontrados', { exact: false })).toHaveTextContent(
      '3 arquivos encontrados'
    );
    expect(screen.getByText('pdf')).toHaveTextContent('pdf 2');
    expect(screen.getByText('Outras')).toHaveTextContent('Outras 1');
  });

  it('ações chamam os bindings e mostram o estado devolvido', async () => {
    const initial = viewState({ sourceFolderPath: 'C:\\origem' });
    const organized = viewState({ sourceFolderPath: 'C:\\origem', hasUndo: true });
    mockBackend(initial, {
      Organize: vi.fn().mockResolvedValue(organized)
    });
    render(<App />);

    const main = await screen.findByRole('main');
    fireEvent.click(within(main).getByRole('button', { name: 'Organizar' }));

    await vi.waitFor(() => expect(window.go.app.App.Organize).toHaveBeenCalled());
    await vi.waitFor(() =>
      expect(within(main).getByRole('button', { name: 'Desfazer' })).toBeEnabled()
    );
  });
});
