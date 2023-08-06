import { useState, useCallback } from "react";
import { Text, TreeView, Box } from "@primer/react";
import { PlusIcon } from "@primer/octicons-react";
import { Dialog } from '@primer/react/drafts'


const GoalsPanel = () => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const openDialog = useCallback(() => setIsOpen(true), [setIsOpen])
    const closeDialog = useCallback(() => setIsOpen(false), [setIsOpen])
    return (
        <nav aria-label="Goals Content">
            <TreeView aria-label="Goals Content">
                <TreeView.Item id="goals-root">
                    <TreeView.LeadingVisual>
                        <TreeView.DirectoryIcon />
                    </TreeView.LeadingVisual>
                    ..
                    <TreeView.SubTree>
                        <TreeView.Item id="create-goal" onSelect={openDialog}>
                            <TreeView.LeadingVisual>
                                <PlusIcon />
                            </TreeView.LeadingVisual>
                            Create Goal
                            {isOpen && (
                                <Dialog
                                    title="Dialog example"
                                    subtitle={
                                        <>
                                            This is a <b>description</b> of the dialog.
                                        </>
                                    }
                                    footerButtons={[{ content: 'Ok', onClick: closeDialog }]}
                                    onClose={closeDialog}
                                >
                                    <Text fontFamily="sans-serif">Some content</Text>
                                </Dialog>
                            )}
                        </TreeView.Item>
                    </TreeView.SubTree>
                </TreeView.Item>
            </TreeView>
        </nav>
    )
}


export default GoalsPanel