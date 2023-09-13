import { Box, FormControl, IconButton, TextInput, Tooltip } from "@primer/react"
import { useVince } from "../../providers";
import { useCallback, useEffect, useState } from "react";
import { Dialog } from '@primer/react/drafts'
import { PlusIcon } from "@primer/octicons-react";



export type CreateSiteDialogProps = {
    afterCreate: () => void;
}

const domainRe = new RegExp("^(?!-)[A-Za-z0-9-]+([-.]{1}[a-z0-9]+)*.[A-Za-z]{2,6}$")

export const CreateSiteDialog = ({ afterCreate }: CreateSiteDialogProps) => {
    const { sitesClient } = useVince()

    const [isOpen, setIsOpen] = useState(false);
    const openDialog = useCallback(() => setIsOpen(true), [setIsOpen])
    const closeDialog = useCallback(() => setIsOpen(false), [setIsOpen])
    const [domain, setDomain] = useState<string>("")

    const submitNewSite = useCallback(() => {
        setIsOpen(false)
        sitesClient?.createSite({ domain }).then((result) => {
            afterCreate()
        })
            .catch((e) => {
                console.log(e)
            })
    }, [domain, afterCreate])

    const [validDomain, setValidDomain] = useState(true)


    useEffect(() => {
        if (domain != "") {
            if (domainRe.test(domain)) {
                setValidDomain(true)
            } else {
                setValidDomain(false)
            }
        }
    }, [domain])

    return (
        <>
            <Tooltip aria-label="Add new site" direction="sw">
                <IconButton aria-label="add site" onClick={openDialog} icon={PlusIcon} />
            </Tooltip>
            {isOpen && <Dialog
                title="Create New Site"
                footerButtons={
                    [{
                        content: 'Create', onClick: submitNewSite,
                    }]
                }
                onClose={closeDialog}
            >
                <Box>
                    <FormControl>
                        <FormControl.Label>Domain</FormControl.Label>
                        <TextInput
                            monospace
                            block
                            placeholder="vinceanalytics.github.io"
                            onChange={(e) => setDomain(e.currentTarget.value)}
                        />
                        {!validDomain &&
                            <FormControl.Validation id="new-site" variant="error">
                                Domain must be the
                            </FormControl.Validation>}
                    </FormControl>
                </Box>
            </Dialog>}
        </>
    )
}