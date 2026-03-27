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
    feedback,
    handleClearFeedback,
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
      feedback={feedback}
      onClearFeedback={handleClearFeedback}
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