import { Box } from "@primer/react"
import { TableIcon, GraphIcon, DownloadIcon } from "@primer/octicons-react";
import { UnderlineNav } from '@primer/react/drafts'
import { useState } from "react";

export const Result = () => {
    const [panel, setPanel] = useState<string>("grid")
    return (
        <Box
            overflow={"hidden"}
        >
            <UnderlineNav aria-label="Results">
                <UnderlineNav.Item
                    icon={TableIcon}
                    aria-current={panel === "grid" ? "page" : undefined}
                    onSelect={() => setPanel("grid")}
                >
                    Grid
                </UnderlineNav.Item>
                <UnderlineNav.Item icon={GraphIcon}
                    aria-current={panel === "graph" ? "page" : undefined}
                    onSelect={() => setPanel("graph")}
                >
                    Graph
                </UnderlineNav.Item>
                <UnderlineNav.Item icon={DownloadIcon}
                >
                    CSV
                </UnderlineNav.Item>
            </UnderlineNav>
            {panel === "grid" && <Grid />}
            {panel === "graph" && <Graph />}
        </Box>
    )
}



const Grid = () => {
    return (
        <Box>
            <h1>GRID</h1>
        </Box>
    )
}

const Graph = () => {
    return (
        <Box>
            <h1>GRAPH</h1>
        </Box>
    )
}