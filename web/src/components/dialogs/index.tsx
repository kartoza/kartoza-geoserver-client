import ConnectionDialog from './ConnectionDialog'
import WorkspaceDialog from './WorkspaceDialog'
import ConfirmDialog from './ConfirmDialog'
import UploadDialog from './UploadDialog'
import LayerGroupDialog from './LayerGroupDialog'
import CacheDialog from './CacheDialog'
import LayerDialog from './LayerDialog'
import StoreDialog from './StoreDialog'
import { SettingsDialog } from './SettingsDialog'
import { SyncDialog } from './SyncDialog'

export default function Dialogs() {
  return (
    <>
      <ConnectionDialog />
      <WorkspaceDialog />
      <ConfirmDialog />
      <UploadDialog />
      <LayerGroupDialog />
      <CacheDialog />
      <LayerDialog />
      <StoreDialog />
      <SyncDialog />
    </>
  )
}

export { SettingsDialog, SyncDialog }
