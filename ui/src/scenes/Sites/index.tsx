import { useState, useCallback, ReactNode, useEffect } from "react";
import {
  Text, TextInput, FormControl,
  TreeView, Box, ButtonGroup, IconButton, Octicon, Spinner, Tooltip,
} from "@primer/react";
import { PlusIcon, DatabaseIcon, ColumnsIcon, GoalIcon, AlertFillIcon } from "@primer/octicons-react";
import { Dialog, PageHeader } from '@primer/react/drafts'
import { columns } from "../Editor/Monaco/sql";
import styled from "styled-components"
import { useSites, useVince } from "../../providers";
import { CreateSiteDialog, DeleteSiteDialog } from "../dialogs";


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
  const { sites, refresh, selectSite } = useSites()
  return (
    <Wrapper>
      <Box
        display={"grid"}
        gridTemplateColumns={"1fr auto"}
        alignItems={"center"}
        borderBottomWidth={1}
        borderBottomStyle={"solid"}
        borderBottomColor={"border.default"}
        paddingX={1}
      >
        <Box>
        </Box>
        <Box>
          <ButtonGroup>
            <CreateSiteDialog afterCreate={refresh} />
            <Tooltip aria-label="Create  new Goal" direction="sw">
              <IconButton aria-label="new goal" icon={GoalIcon} />
            </Tooltip>
            <Tooltip aria-label="Create  new Alert" direction="sw">
              <IconButton aria-label="new alert" icon={AlertFillIcon} />
            </Tooltip>
          </ButtonGroup>
        </Box>

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
                  <TreeView.Item id={site.domain}
                    onSelect={() => {
                      selectSite(site.domain)
                    }}
                  >
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