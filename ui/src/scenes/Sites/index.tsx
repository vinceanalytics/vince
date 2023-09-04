import { useState, useCallback, ReactNode, useEffect } from "react";
import {
  Text, TextInput, FormControl,
  TreeView, Box, themeGet, IconButton, Octicon, Spinner,
} from "@primer/react";
import { PlusIcon, DatabaseIcon, ColumnsIcon } from "@primer/octicons-react";
import { Dialog, PageHeader } from '@primer/react/drafts'
import { columns } from "../Editor/Monaco/sql";
import styled from "styled-components"
import { useSites, useVince } from "../../providers";


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

type ColumnProps = {
  id: string,
}
const Columns = ({ id }: ColumnProps) => {
  return (
    <TreeView.Item id={`${id}-columns`} >
      <TreeView.LeadingVisual>
        <TreeView.DirectoryIcon />
      </TreeView.LeadingVisual>
      columns
      <TreeView.SubTree>
        {columns.map((name) => (
          <TreeView.Item id={`${id}-columns${name}`}>
            <TreeView.LeadingVisual>
              <Octicon icon={ColumnsIcon} />
            </TreeView.LeadingVisual>
            {name}
          </TreeView.Item>
        ))}
      </TreeView.SubTree>
    </TreeView.Item >
  )
}

const Sites = () => {
  const { vince } = useVince()
  const { sites, refresh } = useSites()

  const [isOpen, setIsOpen] = useState(false);
  const openDialog = useCallback(() => setIsOpen(true), [setIsOpen])
  const closeDialog = useCallback(() => setIsOpen(false), [setIsOpen])
  const [domain, setDomain] = useState<string>("")

  const submitNewSite = useCallback(() => {
    setIsOpen(false)
    vince?.createSite({ domain }).then((result) => {
      refresh()
    })
      .catch((e) => {
        console.log(e)
      })
  }, [domain, refresh])

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
    <Wrapper>
      <Box paddingX={2} sx={{ borderBottomWidth: 1, borderBottomStyle: 'solid', borderColor: 'border.default', pb: 1 }}>
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

        {sites.length == 0 &&
          <Box sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: "100%",
            pt: 2,
          }}>
            <Text>No Sites</Text>
          </Box>}
        {sites.length !== 0 &&
          <Box sx={{
            width: "100%",
            pt: 2,
            overflow: "auto",
          }}>
            <nav>
              <TreeView aria-label="Sites">
                {sites?.map((site) => (
                  <TreeView.Item id={site.domain}>
                    <TreeView.LeadingVisual>
                      <TreeView.DirectoryIcon />
                    </TreeView.LeadingVisual>
                    {site.domain}
                    <TreeView.SubTree>
                      <Columns id={site.domain} />
                    </TreeView.SubTree>
                  </TreeView.Item>
                ))}
              </TreeView>
            </nav>
          </Box>}
      </Box>
    </Wrapper >
  )
}

export default Sites;