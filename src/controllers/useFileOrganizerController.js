import { useEffect, useState } from 'react';

const feedbackByLanguage = {
  'pt-BR': {
    sourceSelectError: 'Não foi possível selecionar a pasta de origem.',
    destinationSelectError: 'Não foi possível selecionar a pasta de destino.',
    sourceRequired: 'Selecione a pasta de origem antes de organizar.',
    organizeUnexpectedError: 'Erro inesperado ao organizar os arquivos.',
    undoUnexpectedError: 'Erro inesperado ao desfazer a organização.',
    droppedPathUnexpectedError: 'Não foi possível usar o item arrastado.',
    droppedPathSuccess: 'Origem definida por arrastar e soltar.',
    organizeSuccess: (result) =>
      `Organizacao concluida: ${result.movedFiles} arquivo(s) movido(s). Origem: ${result.sourceFolderPath}. Destino: ${result.destinationFolderPath}. Processados: ${result.processedFiles}. Ignorados sem extensao: ${result.ignoredWithoutExtension}. Pastas ignoradas: ${result.ignoredFolders}.`,
    undoSuccess: (result) =>
      `Desfazer concluido: ${result.restoredFiles} arquivo(s) restaurado(s). Renomeados na restauracao: ${result.renamedOnRestore}. Nao encontrados: ${result.skippedMissing}.`
  },
  en: {
    sourceSelectError: 'Could not select the source folder.',
    destinationSelectError: 'Could not select the destination folder.',
    sourceRequired: 'Select a source folder before organizing.',
    organizeUnexpectedError: 'Unexpected error while organizing files.',
    undoUnexpectedError: 'Unexpected error while undoing organization.',
    droppedPathUnexpectedError: 'Could not use the dropped item.',
    droppedPathSuccess: 'Source folder set from drag and drop.',
    organizeSuccess: (result) =>
      `Organization complete: ${result.movedFiles} file(s) moved. Source: ${result.sourceFolderPath}. Destination: ${result.destinationFolderPath}. Processed: ${result.processedFiles}. Ignored without extension: ${result.ignoredWithoutExtension}. Ignored folders: ${result.ignoredFolders}.`,
    undoSuccess: (result) =>
      `Undo complete: ${result.restoredFiles} file(s) restored. Renamed on restore: ${result.renamedOnRestore}. Missing: ${result.skippedMissing}.`
  }
};

function useFileOrganizerController(language, organizationOptions) {
  const copy = feedbackByLanguage[language] || feedbackByLanguage['pt-BR'];
  const [sourceFolderPath, setSourceFolderPath] = useState('');
  const [destinationFolderPath, setDestinationFolderPath] = useState('');
  const [hasUndo, setHasUndo] = useState(false);
  const [loadingAction, setLoadingAction] = useState(null);
  const [feedback, setFeedback] = useState(null);

  const isLoading = Boolean(loadingAction);

  const emitFeedback = (type, message) => {
    setFeedback({
      id: `${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
      type,
      message
    });
  };

  useEffect(() => {
    let isMounted = true;

    const loadLastOrganizationState = async () => {
      try {
        const result = await window.electronAPI.getLastOrganizationState();

        if (isMounted) {
          setHasUndo(Boolean(result?.hasUndo));

          if (result?.sourceFolderPath) {
            setSourceFolderPath(result.sourceFolderPath);
          }

          if (result?.destinationFolderPath) {
            setDestinationFolderPath(result.destinationFolderPath);
          }
        }
      } catch {
        if (isMounted) {
          setHasUndo(false);
        }
      }
    };

    loadLastOrganizationState();

    return () => {
      isMounted = false;
    };
  }, []);

  const handleSelectSourceFolder = async () => {
    try {
      const selectedPath = await window.electronAPI.selectSourceFolder();
      if (!selectedPath) {
        return;
      }

      setSourceFolderPath(selectedPath);
      setFeedback(null);
    } catch {
      emitFeedback('error', copy.sourceSelectError);
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
      emitFeedback('error', copy.destinationSelectError);
    }
  };

  const handleResolveDroppedPath = async (droppedPath) => {
    try {
      const result = await window.electronAPI.resolveDroppedPath(droppedPath);
      if (!result?.sourceFolderPath) {
        return;
      }

      setSourceFolderPath(result.sourceFolderPath);

      emitFeedback('info', `${copy.droppedPathSuccess} ${result.sourceFolderPath}`);
    } catch (error) {
      emitFeedback('error', error?.message || copy.droppedPathUnexpectedError);
    }
  };

  const handleOrganizeFiles = async () => {
    if (!sourceFolderPath) {
      emitFeedback('error', copy.sourceRequired);
      return;
    }

    setLoadingAction('organize');
    setFeedback(null);

    try {
      const result = await window.electronAPI.organizeFiles({
        sourceFolderPath,
        destinationFolderPath: destinationFolderPath || sourceFolderPath,
        organizationOptions
      });

      setHasUndo(result.canUndo);

      emitFeedback('organize', copy.organizeSuccess(result));
    } catch (error) {
      emitFeedback('error', error?.message || copy.organizeUnexpectedError);
    } finally {
      setLoadingAction(null);
    }
  };

  const handleUndoLastOrganization = async () => {
    setLoadingAction('restore');
    setFeedback(null);

    try {
      const result = await window.electronAPI.undoLastOrganization();
      setHasUndo(result.canUndo);

      emitFeedback('restore', copy.undoSuccess(result));
    } catch (error) {
      emitFeedback('error', error?.message || copy.undoUnexpectedError);
    } finally {
      setLoadingAction(null);
    }
  };

  const handleClearFeedback = () => {
    setFeedback(null);
  };

  return {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    loadingAction,
    feedback,
    handleClearFeedback,
    handleResolveDroppedPath,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization
  };
}

export default useFileOrganizerController;