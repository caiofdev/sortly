import useFileOrganizerController from './controllers/useFileOrganizerController';
import useLanguagePreference from './hooks/useLanguagePreference';
import OrganizerView from './views/OrganizerView';

function App() {
  const { language, setLanguage } = useLanguagePreference();

  const {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    loadingAction,
    feedback,
    handleResolveDroppedPath,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization
  } = useFileOrganizerController(language);

  return (
    <OrganizerView
      language={language}
      sourceFolderPath={sourceFolderPath}
      destinationFolderPath={destinationFolderPath}
      hasUndo={hasUndo}
      isLoading={isLoading}
      loadingAction={loadingAction}
      feedback={feedback}
      onLanguageChange={setLanguage}
      onResolveDroppedPath={handleResolveDroppedPath}
      onSelectSourceFolder={handleSelectSourceFolder}
      onSelectDestinationFolder={handleSelectDestinationFolder}
      onOrganizeFiles={handleOrganizeFiles}
      onUndoLastOrganization={handleUndoLastOrganization}
    />
  );
}

export default App;