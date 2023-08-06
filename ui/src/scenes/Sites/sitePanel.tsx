import { useState, useCallback } from "react";
import { Text, TreeView, Box, Button, IconButton, Link } from "@primer/react";
import { PlusIcon } from "@primer/octicons-react";
import { Dialog } from '@primer/react/drafts'


const SitePanel = () => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const openDialog = useCallback(() => setIsOpen(true), [setIsOpen])
    const closeDialog = useCallback(() => setIsOpen(false), [setIsOpen])
    return (
        <nav aria-label="Sites Content">
            <TreeView aria-label="Sites Content">
                <TreeView.Item id="root">
                    <TreeView.LeadingVisual>
                        <TreeView.DirectoryIcon />
                    </TreeView.LeadingVisual>
                    ..
                    <TreeView.SubTree>
                        <TreeView.Item id="createSite" >
                            <TreeView.LeadingVisual>
                                <PlusIcon />
                            </TreeView.LeadingVisual>
                            <Link onClick={openDialog}>
                                Create Site
                            </Link>
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


export default SitePanel