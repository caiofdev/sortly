import appLogo from '../assets/app-logo.svg';
import DragDropPanel from '../components/DragDropPanel';
import LanguageToggle from '../components/LanguageToggle';

const copy = {
  'pt-BR': {
    subtitle: 'Selecione origem e destino, clique em organizar e pronto.',
    sourceButton: 'Selecionar pasta de origem',
    destinationButton: 'Selecionar pasta de destino',
    sourceLabel: 'Origem',
    destinationLabel: 'Destino',
    sourceEmpty: 'Nenhuma pasta de origem selecionada',
    destinationEmpty: 'Nenhuma pasta de destino selecionada',
    organize: 'Organizar arquivos',
    organizing: 'Organizando arquivos...',
    undo: 'Desfazer ultima separacao',
    feedbackEmpty: 'Nenhuma acao executada ainda. Selecione as pastas e clique em organizar.',
    dropTitle: 'Arraste e solte uma pasta ou arquivo aqui',
    dropDescription: 'Se for arquivo, usamos a pasta dele como origem.',
    dropSelectHintPrefix: 'Ou',
    dropSelectHintAction: 'selecione uma pasta'
  },
  en: {
    subtitle: 'Select source and destination, click organize, and done.',
    sourceButton: 'Select source folder',
    destinationButton: 'Select destination folder',
    sourceLabel: 'Source',
    destinationLabel: 'Destination',
    sourceEmpty: 'No source folder selected',
    destinationEmpty: 'No destination folder selected',
    organize: 'Organize files',
    organizing: 'Organizing files...',
    undo: 'Undo last organization',
    feedbackEmpty: 'No action yet. Select folders and click organize.',
    dropTitle: 'Drag and drop a folder or file here',
    dropDescription: 'If a file is dropped, we use its parent folder as source.',
    dropSelectHintPrefix: 'Or',
    dropSelectHintAction: 'select a folder'
  }
};

function OrganizerView({
  language,
  sourceFolderPath,
  destinationFolderPath,
  hasUndo,
  isLoading,
  feedback,
  onLanguageChange,
  onResolveDroppedPath,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  const text = copy[language] || copy['pt-BR'];
  const destinationLabel = destinationFolderPath || sourceFolderPath || text.destinationEmpty;

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#0F172A] text-[#F8FAFC]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_10%_15%,rgba(59,130,246,0.16),transparent_35%),radial-gradient(circle_at_90%_85%,rgba(34,197,94,0.12),transparent_30%)]" />

      <section className="relative mx-auto flex min-h-screen w-full max-w-4xl items-center justify-center px-6 py-12">
        <div className="w-full rounded-2xl border border-white/10 bg-[#1E293B]/90 p-8 shadow-[0_22px_60px_rgba(2,6,23,0.45)] backdrop-blur-sm md:p-10">
          <header className="space-y-3 text-center">
            <div className="flex items-center justify-between gap-3 pb-2">
              <LanguageToggle language={language} onChange={onLanguageChange} />
            </div>

            <div className="flex items-center justify-center gap-3">
              <img src={appLogo} alt="Logo do app" className="h-14 w-14" />
              <h1 className="text-3xl font-bold tracking-tight md:text-4xl">Sortly</h1>
            </div>
            <p className="text-sm text-[#94A3B8] md:text-base">{text.subtitle}</p>
          </header>

          <div className="mt-8 grid gap-5">
            <DragDropPanel
              isLoading={isLoading}
              labels={text}
              onResolveDroppedPath={onResolveDroppedPath}
              onSelectSourceFolder={onSelectSourceFolder}
            />

            <button
              type="button"
              onClick={onSelectDestinationFolder}
              className="rounded-2xl bg-[#334155] px-6 py-4 text-base font-semibold text-[#F8FAFC] shadow-sm transition-all hover:-translate-y-0.5 hover:bg-[#3b4a5d] hover:shadow-md"
            >
              {text.destinationButton}
            </button>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="rounded-2xl border border-white/10 bg-[#0F172A]/65 px-4 py-4 shadow-sm">
                <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-[#94A3B8]">{text.sourceLabel}</p>
                <p className="min-h-12 break-all text-sm text-[#F8FAFC]">{sourceFolderPath || text.sourceEmpty}</p>
              </div>

              <div className="rounded-2xl border border-white/10 bg-[#0F172A]/65 px-4 py-4 shadow-sm">
                <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-[#94A3B8]">{text.destinationLabel}</p>
                <p className="min-h-12 break-all text-sm text-[#F8FAFC]">{destinationLabel}</p>
              </div>
            </div>

            <button
              type="button"
              onClick={onOrganizeFiles}
              disabled={isLoading || !sourceFolderPath}
              className="rounded-2xl bg-[#3B82F6] px-6 py-4 text-lg font-semibold text-white shadow-[0_10px_30px_rgba(59,130,246,0.35)] transition-all hover:-translate-y-0.5 hover:bg-[#4f8ff7] disabled:cursor-not-allowed disabled:bg-[#2b3b57] disabled:text-slate-300 disabled:shadow-none"
            >
              {isLoading ? text.organizing : text.organize}
            </button>

            <button
              type="button"
              onClick={onUndoLastOrganization}
              disabled={isLoading || !hasUndo}
              className="rounded-2xl border border-[#3B82F6]/45 bg-transparent px-6 py-4 text-base font-semibold text-[#F8FAFC] transition-all hover:-translate-y-0.5 hover:border-[#3B82F6] hover:bg-[#3B82F6]/12 disabled:cursor-not-allowed disabled:border-slate-600/60 disabled:text-slate-400"
            >
              {text.undo}
            </button>

            <div
              className={`rounded-2xl border px-4 py-4 text-sm shadow-sm ${
                feedback
                  ? feedback.type === 'success'
                    ? 'border-[#22C55E]/45 bg-[#22C55E]/12 text-[#b8ffcf]'
                    : 'border-rose-500/45 bg-rose-500/12 text-rose-200'
                  : 'border-white/10 bg-[#0F172A]/65 text-[#94A3B8]'
              }`}
            >
              {feedback ? feedback.message : text.feedbackEmpty}
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}

export default OrganizerView;