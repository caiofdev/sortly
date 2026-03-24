import { useState } from 'react';

function App() {
  const [sourceFolderPath, setSourceFolderPath] = useState('');
  const [destinationFolderPath, setDestinationFolderPath] = useState('');
  const [hasUndo, setHasUndo] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [feedback, setFeedback] = useState(null);

  const handleSelectSourceFolder = async () => {
    try {
      const selectedPath = await window.electronAPI.selectSourceFolder();
      if (!selectedPath) {
        return;
      }

      setSourceFolderPath(selectedPath);
      if (!destinationFolderPath) {
        setDestinationFolderPath(selectedPath);
      }
      setFeedback(null);
    } catch (error) {
      setFeedback({
        type: 'error',
        message: 'Não foi possível selecionar a pasta de origem.'
      });
    }
  };

  const handleSelectDestinationFolder = async () => {
    try {
      const selectedPath = await window.electronAPI.selectDestinationFolder();
      if (!selectedPath) {
        return;
      }

      setDestinationFolderPath(selectedPath);
      setFeedback(null);
    } catch (error) {
      setFeedback({
        type: 'error',
        message: 'Não foi possível selecionar a pasta de destino.'
      });
    }
  };

  const handleOrganizeFiles = async () => {
    if (!sourceFolderPath) {
      setFeedback({
        type: 'error',
        message: 'Selecione a pasta de origem antes de organizar.'
      });
      return;
    }

    setIsLoading(true);
    setFeedback(null);

    try {
      const result = await window.electronAPI.organizeFiles({
        sourceFolderPath,
        destinationFolderPath: destinationFolderPath || sourceFolderPath
      });

      setHasUndo(result.canUndo);

      setFeedback({
        type: 'success',
        message: `${result.message} Origem: ${result.sourceFolderPath}. Destino: ${result.destinationFolderPath}. Processados: ${result.processedFiles}. Ignorados sem extensão: ${result.ignoredWithoutExtension}. Pastas ignoradas: ${result.ignoredFolders}.`
      });
    } catch (error) {
      setFeedback({
        type: 'error',
        message: error?.message || 'Erro inesperado ao organizar os arquivos.'
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleUndoLastOrganization = async () => {
    setIsLoading(true);
    setFeedback(null);

    try {
      const result = await window.electronAPI.undoLastOrganization();
      setHasUndo(result.canUndo);

      setFeedback({
        type: 'success',
        message: `${result.message} Renomeados na restauração: ${result.renamedOnRestore}. Não encontrados: ${result.skippedMissing}.`
      });
    } catch (error) {
      setFeedback({
        type: 'error',
        message: error?.message || 'Erro inesperado ao desfazer a organização.'
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center p-6">
      <section className="w-full max-w-2xl rounded-2xl border border-slate-800 bg-slate-900/90 p-8 shadow-2xl">
        <h1 className="text-3xl font-bold tracking-tight text-center">Organizador de Arquivos</h1>

        <div className="mt-8 space-y-4">
          <button
            type="button"
            onClick={handleSelectSourceFolder}
            className="w-full rounded-xl bg-slate-700 hover:bg-slate-600 transition-colors px-5 py-4 text-lg font-semibold"
          >
            Selecionar pasta de origem
          </button>

          <div className="rounded-xl border border-slate-700 bg-slate-950/70 px-4 py-3 text-sm text-slate-300 break-all min-h-14">
            <p className="text-xs uppercase tracking-wide text-slate-500 mb-1">Origem</p>
            <p>{sourceFolderPath || 'Nenhuma pasta de origem selecionada'}</p>
          </div>

          <button
            type="button"
            onClick={handleSelectDestinationFolder}
            className="w-full rounded-xl bg-slate-700 hover:bg-slate-600 transition-colors px-5 py-4 text-lg font-semibold"
          >
            Selecionar pasta de destino
          </button>

          <div className="rounded-xl border border-slate-700 bg-slate-950/70 px-4 py-3 text-sm text-slate-300 break-all min-h-14">
            <p className="text-xs uppercase tracking-wide text-slate-500 mb-1">Destino</p>
            <p>{destinationFolderPath || sourceFolderPath || 'Nenhuma pasta de destino selecionada'}</p>
          </div>

          <button
            type="button"
            onClick={handleOrganizeFiles}
            disabled={isLoading || !sourceFolderPath}
            className="w-full rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:bg-indigo-800/70 disabled:cursor-not-allowed transition-colors px-5 py-4 text-lg font-semibold"
          >
            {isLoading ? 'Organizando arquivos...' : 'Organizar arquivos'}
          </button>

          <button
            type="button"
            onClick={handleUndoLastOrganization}
            disabled={isLoading || !hasUndo}
            className="w-full rounded-xl bg-amber-600 hover:bg-amber-500 disabled:bg-amber-900/50 disabled:cursor-not-allowed transition-colors px-5 py-4 text-lg font-semibold"
          >
            Desfazer última separação
          </button>

          <div
            className={`rounded-xl border px-4 py-3 text-sm min-h-14 flex items-center ${
              feedback
                ? feedback.type === 'success'
                  ? 'border-emerald-500/40 bg-emerald-900/30 text-emerald-300'
                  : 'border-rose-500/40 bg-rose-900/30 text-rose-300'
                : 'border-slate-700 bg-slate-950/50 text-slate-500'
            }`}
          >
            {feedback ? feedback.message : 'O resultado da organização aparecerá aqui.'}
          </div>
        </div>
      </section>
    </main>
  );
}

export default App;
