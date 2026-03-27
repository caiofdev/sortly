import { useState } from 'react';
import useFileOrganizerController from './controllers/useFileOrganizerController';
import useLanguagePreference from './hooks/useLanguagePreference';
import OrganizerView from './views/OrganizerView';

function App() {
  const { language, setLanguage } = useLanguagePreference();
  const [viewMode, setViewMode] = useState('default');

  const {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
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
      viewMode={viewMode}
      sourceFolderPath={sourceFolderPath}
      destinationFolderPath={destinationFolderPath}
      hasUndo={hasUndo}
      isLoading={isLoading}
      feedback={feedback}
      onLanguageChange={setLanguage}
      onViewModeChange={setViewMode}
      onResolveDroppedPath={handleResolveDroppedPath}
      onSelectSourceFolder={handleSelectSourceFolder}
      onSelectDestinationFolder={handleSelectDestinationFolder}
      onOrganizeFiles={handleOrganizeFiles}
      onUndoLastOrganization={handleUndoLastOrganization}
    />
  );
}

export default App;