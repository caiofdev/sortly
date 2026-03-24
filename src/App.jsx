import { useState } from 'react';

function App() {
  const [folderPath, setFolderPath] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [feedback, setFeedback] = useState(null);

  const handleSelectFolder = async () => {
    try {
      const selectedPath = await window.electronAPI.selectFolder();
      if (!selectedPath) {
        return;
      }

      setFolderPath(selectedPath);
      setFeedback(null);
    } catch (error) {
      setFeedback({
        type: 'error',
        message: 'Não foi possível selecionar a pasta.'
      });
    }
  };

  const handleOrganizeFiles = async () => {
    if (!folderPath) {
      setFeedback({
        type: 'error',
        message: 'Selecione uma pasta antes de organizar.'
      });
      return;
    }

    setIsLoading(true);
    setFeedback(null);

    try {
      const result = await window.electronAPI.organizeFiles(folderPath);
      setFeedback({
        type: 'success',
        message: `${result.message} Processados: ${result.processedFiles}. Ignorados sem extensão: ${result.ignoredWithoutExtension}. Pastas ignoradas: ${result.ignoredFolders}.`
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

  return (
    <main className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center p-6">
      <section className="w-full max-w-2xl rounded-2xl border border-slate-800 bg-slate-900/90 p-8 shadow-2xl">
        <h1 className="text-3xl font-bold tracking-tight text-center">Organizador de Arquivos</h1>

        <div className="mt-8 space-y-4">
          <button
            type="button"
            onClick={handleSelectFolder}
            className="w-full rounded-xl bg-slate-700 hover:bg-slate-600 transition-colors px-5 py-4 text-lg font-semibold"
          >
            Selecionar pasta
          </button>

          <div className="rounded-xl border border-slate-700 bg-slate-950/70 px-4 py-3 text-sm text-slate-300 break-all min-h-14 flex items-center">
            {folderPath || 'Nenhuma pasta selecionada'}
          </div>

          <button
            type="button"
            onClick={handleOrganizeFiles}
            disabled={isLoading || !folderPath}
            className="w-full rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:bg-indigo-800/70 disabled:cursor-not-allowed transition-colors px-5 py-4 text-lg font-semibold"
          >
            {isLoading ? 'Organizando arquivos...' : 'Organizar arquivos'}
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
