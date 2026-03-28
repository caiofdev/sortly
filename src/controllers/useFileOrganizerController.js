import { useEffect, useState } from 'react';
import feedbackCopy from '../i18n/feedbackCopy';

function useFileOrganizerController(language, organizationOptions) {
  const copy = feedbackCopy[language] || feedbackCopy['pt-BR'];
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
        const hasUndo = Boolean(result?.hasUndo);

        if (isMounted) {
          setHasUndo(hasUndo);

          if (result?.sourceFolderPath) {
            setSourceFolderPath(result.sourceFolderPath);
          }

          if (result?.destinationFolderPath) {
            setDestinationFolderPath(result.destinationFolderPath);
          }

          if (hasUndo) {
            emitFeedback('info', copy.recoveredLastOrganization);
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