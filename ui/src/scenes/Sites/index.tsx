import { useState, useCallback, ReactNode, useEffect } from "react";
import {
  Text, TextInput, FormControl,
  TreeView, Box, ButtonGroup, IconButton, Octicon, Spinner, Tooltip, Label,
} from "@primer/react";
import { PlusIcon, DatabaseIcon, ColumnsIcon, GoalIcon, AlertFillIcon } from "@primer/octicons-react";
import { Dialog, PageHeader } from '@primer/react/drafts'
import { columns } from "../Editor/Monaco/sql";
import styled from "styled-components"
import { useSites, useVince } from "../../providers";
import { CreateSiteDialog, DeleteSiteDialog } from "../dialogs";
import { CreateGoalDialog } from "../dialogs/goals";
import { Goal, Goal_Type, Site } from "../../vince";


export const PaneWrapper = styled.div`
  display: flex;
  flex-direction: column;
  flex: 1;
`

const Wrapper = styled(PaneWrapper)`
  overflow-x: auto;
  height: 100%;
`





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
        <Box
          display={"grid"}
          gridTemplateColumns={"auto auto auto"}
          sx={{ gap: "1" }}
        >
          <Box>
            <CreateSiteDialog afterCreate={refresh} />
          </Box>
          <Box>
            <CreateGoalDialog afterCreate={refresh} />
          </Box>
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
                    <TreeView.TrailingVisual>
                      <Label>site</Label>
                    </TreeView.TrailingVisual>
                    <TreeView.SubTree>
                      <Columns id={site.domain} />
                      <Goals site={site} />
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

const Goals = ({ site }: { site: Site }) => {
  let goals: Goal[] = []
  for (const key in site.goals) {
    goals.push(site.goals[key])
  }
  return (
    <TreeView.Item id={`${site.domain}-goals`}>
      <TreeView.LeadingVisual>
        <TreeView.DirectoryIcon />
      </TreeView.LeadingVisual>
      goals
      <TreeView.SubTree>
        {goals.map((goal) => (
          <TreeView.Item id={`${goal.name}`}>
            {goal.name}
            <TreeView.TrailingVisual>
              <Label>{goal.type === Goal_Type.EVENT ? "custom" : "path"}</Label>
            </TreeView.TrailingVisual>
          </TreeView.Item>
        ))}
      </TreeView.SubTree>
    </TreeView.Item>
  )
}

export default Sites;