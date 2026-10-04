import { useEffect, useRef, useState } from 'react';
import { describeError } from '../i18n/describeError';
import feedbackCopy from '../i18n/feedbackCopy';
import { getCopy } from '../i18n/language';
import defaultGateway from '../services/sortlyGateway';

function useFileOrganizerController({
  language,
  notify,
  setLanguage,
  setCriterion,
  gateway = defaultGateway
}) {
  const copy = getCopy(feedbackCopy, language);
  const copyRef = useRef(copy);
  copyRef.current = copy;

  const [sourceFolderPath, setSourceFolderPath] = useState('');
  const [destinationFolderPath, setDestinationFolderPath] = useState('');
  const [hasUndo, setHasUndo] = useState(false);
  const [loadingAction, setLoadingAction] = useState(null);

  useEffect(() => {
    let isMounted = true;

    gateway
      .getLastOrganizationState()
      .then((state) => {
        if (!isMounted) return;
        setHasUndo(Boolean(state?.hasUndo));
        if (state?.sourceFolderPath) setSourceFolderPath(state.sourceFolderPath);
        if (state?.destinationFolderPath) setDestinationFolderPath(state.destinationFolderPath);
        if (state?.hasUndo) notify('info', copyRef.current.recoveredLastOrganization);
      })
      .catch(() => {
        if (isMounted) setHasUndo(false);
      });

    return () => {
      isMounted = false;
    };
  }, [gateway, notify]);

  const selectFolder = async (pick, setPath, errorMessage) => {
    try {
      const selectedPath = await pick();
      if (selectedPath) setPath(selectedPath);
    } catch {
      notify('error', errorMessage);
    }
  };

  const runAction = async (kind, call, successMessage, fallbackMessage) => {
    setLoadingAction(kind);
    try {
      const result = await call();
      setHasUndo(Boolean(result.canUndo));
      notify(kind, successMessage(result));
    } catch (error) {
      notify('error', describeError(error, copy, fallbackMessage));
    } finally {
      setLoadingAction(null);
    }
  };

  const handleSelectSourceFolder = () =>
    selectFolder(gateway.selectSourceFolder, setSourceFolderPath, copy.sourceSelectError);

  const handleSelectDestinationFolder = () =>
    selectFolder(
      gateway.selectDestinationFolder,
      setDestinationFolderPath,
      copy.destinationSelectError
    );

  const handleResolveDroppedPath = async (droppedPath) => {
    try {
      const result = await gateway.resolveDroppedPath(droppedPath);
      if (!result?.sourceFolderPath) return;
      setSourceFolderPath(result.sourceFolderPath);
      notify('info', `${copy.droppedPathSuccess} ${result.sourceFolderPath}`);
    } catch (error) {
      notify('error', describeError(error, copy, copy.droppedPathUnexpectedError));
    }
  };

  // Vários itens soltos de uma vez: usa o primeiro.
  const dropRef = useRef(handleResolveDroppedPath);
  dropRef.current = handleResolveDroppedPath;

  useEffect(
    () =>
      gateway.subscribeFileDrop((paths) => {
        if (paths?.length) dropRef.current(paths[0]);
      }),
    [gateway]
  );

  const handleOrganizeFiles = () => {
    if (!sourceFolderPath) {
      notify('error', copy.sourceRequired);
      return Promise.resolve();
    }
    const destination = destinationFolderPath || sourceFolderPath;
    return runAction(
      'organize',
      () => gateway.organizeFiles(sourceFolderPath, destination),
      copy.organizeSuccess,
      copy.organizeUnexpectedError
    );
  };

  const handleUndoLastOrganization = () =>
    runAction('restore', gateway.undoLastOrganization, copy.undoSuccess, copy.undoUnexpectedError);

  const changeSetting = async (change) => {
    try {
      await change();
    } catch (error) {
      notify('error', describeError(error, copy, copy.settingsSaveError));
    }
  };

  const handleLanguageChange = (next) => changeSetting(() => setLanguage(next));

  const handleCriterionChange = (key, enabled) => changeSetting(() => setCriterion(key, enabled));

  return {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading: Boolean(loadingAction),
    loadingAction,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization,
    handleLanguageChange,
    handleCriterionChange
  };
}

export default useFileOrganizerController;
