import useFileOrganizerController from './controllers/useFileOrganizerController';
import OrganizerView from './views/OrganizerView';

function App() {
  const {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    feedback,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization
  } = useFileOrganizerController();

  return (
    <OrganizerView
      sourceFolderPath={sourceFolderPath}
      destinationFolderPath={destinationFolderPath}
      hasUndo={hasUndo}
      isLoading={isLoading}
      feedback={feedback}
      onSelectSourceFolder={handleSelectSourceFolder}
      onSelectDestinationFolder={handleSelectDestinationFolder}
      onOrganizeFiles={handleOrganizeFiles}
      onUndoLastOrganization={handleUndoLastOrganization}
    />
  );
}

export default App;
