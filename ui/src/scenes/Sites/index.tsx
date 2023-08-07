import { useState, useCallback, ReactNode, useEffect } from "react";
import {
  Text, TextInput, FormControl,
  TreeView, Box, themeGet, IconButton, Octicon, Spinner,
} from "@primer/react";
import { PlusIcon, DatabaseIcon } from "@primer/octicons-react";
import { Dialog, PageHeader } from '@primer/react/drafts'

import styled, { css } from "styled-components"
import { Site, Client } from "../../vince";


export const PaneWrapper = styled.div`
  display: flex;
  flex-direction: column;
  flex: 1;
`

const Wrapper = styled(PaneWrapper)`
  overflow-x: auto;
  height: 100%;
`








const domainRe = new RegExp("^(?!-)[A-Za-z0-9-]+([-.]{1}[a-z0-9]+)*.[A-Za-z]{2,6}$")

const Sites = () => {
  const [loading, setLoading] = useState(false)
  const [vince] = useState(new Client())
  const [sites, setSites] = useState<Site[]>()

  const [isOpen, setIsOpen] = useState(false);
  const openDialog = useCallback(() => setIsOpen(true), [setIsOpen])
  const closeDialog = useCallback(() => setIsOpen(false), [setIsOpen])
  const [domain, setDomain] = useState<string>("")

  const submitNewSite = useCallback(() => {
    console.log(domain)
    setIsOpen(false)
  }, [domain])
  const [validDomain, setValidDomain] = useState(true)

  const fetchSites = useCallback(() => {
    setLoading(true)
    vince.sites().then((result) => {
      setSites(result as Site[])
      setLoading(false);
    })
      .catch((e) => {
        console.log(e)
      })
  }, [setLoading, setSites])

  useEffect(() => {
    if (domain != "") {
      if (domainRe.test(domain)) {
        setValidDomain(true)
      } else {
        setValidDomain(false)
      }
    }
  }, [domain])

  useEffect(() => {
    fetchSites()
  }, [fetchSites])

  return (
    <Wrapper>
      <Box paddingX={2}>
        <PageHeader>
          <PageHeader.TitleArea>
            <PageHeader.LeadingAction>
              <DatabaseIcon />
            </PageHeader.LeadingAction>
            <PageHeader.Title>Sites</PageHeader.Title>
            <PageHeader.Actions>
              <IconButton variant="primary" aria-label="New Site" icon={PlusIcon} onClick={openDialog} />
              {isOpen && (
                <Dialog
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
                </Dialog>
              )}
            </PageHeader.Actions>
          </PageHeader.TitleArea>
        </PageHeader>
      </Box>
      <Box display={"flex"} overflow={"auto"}>
        {loading &&
          <Box sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: "100%",
          }}>
            <Spinner size="large" />
          </Box>}
      </Box>
    </Wrapper >
  )
}

export default Sites;