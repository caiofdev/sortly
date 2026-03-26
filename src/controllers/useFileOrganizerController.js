import { useState } from 'react';

function useFileOrganizerController() {
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
    } catch {
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
    } catch {
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

  return {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    feedback,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization
  };
}

export default useFileOrganizerController;