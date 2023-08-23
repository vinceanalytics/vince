import { Box } from "@primer/react"
import { TableIcon, GraphIcon, DownloadIcon } from "@primer/octicons-react";
import { DataTable, UnderlineNav } from '@primer/react/drafts'
import { useState } from "react";
import { useQuery } from "../../providers";
import { QueryResult, Value } from "../../vince";

export const Result = () => {
    const [panel, setPanel] = useState<string>("grid")
    const { result } = useQuery()
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
            {panel === "grid" && <Grid result={result} />}
            {panel === "graph" && <Graph />}
        </Box>
    )
}









const Grid = ({ result }: { result: QueryResult }) => {
    const data = result.rows ? (result.rows.map((row, id) => (
        row.values.reduce((a, v, idx) => ({
            ...a, ...Object.fromEntries([
                [result.columns[idx].name, v]
            ])
        }), { id })
    ))) : [];
    return (
        <Box marginTop={1} height={"100%"}>
            {result && <DataTable
                columns={result.columns.map(({ name }, idx) => ({
                    id: idx.toString(),
                    header: name,
                }))}
                data={data}
            />}
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