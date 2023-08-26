import { Box, Label, Text } from "@primer/react"
import { TableIcon, GraphIcon, DownloadIcon } from "@primer/octicons-react";
import { DataTable, UnderlineNav } from '@primer/react/drafts'
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from "react";
import { useQuery } from "../../providers";
import { QueryResult, Value } from "../../vince";
import { Chart } from "frappe-charts/dist/frappe-charts.min.esm";

export const Result = () => {
    const [panel, setPanel] = useState<string>("grid")
    const { result } = useQuery()
    const chartRef = useRef()

    const exportChart = () => {
        if (chartRef && chartRef.current) {
            //@ts-ignore
            chartRef.current.export();
        }
    };
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
                <UnderlineNav.Item icon={DownloadIcon}>
                    CSV
                </UnderlineNav.Item>
                <UnderlineNav.Item icon={DownloadIcon} onSelect={exportChart}>
                    SVG
                </UnderlineNav.Item>
            </UnderlineNav>
            {panel === "grid" && <Grid result={result} />}
            {panel === "graph" && <FChart
                ref={chartRef}
                type="line"
                data={{
                    labels: labels,
                    datasets: [{ values: [18, 40, 30, 35, 8, 52, 17, 4] }],
                }}
                colors={["#21ba45", "#98d85b"]}
                axisOptions={{
                    xAxisMode: "tick",
                    yAxisMode: "tick",
                    xIsSeries: 1
                }}
            />}
        </Box >
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
        <Box marginTop={1} overflow={"auto"}>
            {result && <DataTable
                columns={result.columns.map(({ name }, idx) => ({
                    id: idx.toString(),
                    header: name,
                    renderCell(data) {
                        //@ts-ignore
                        const value = data[name] as Value;
                        let format = ''
                        if (value.number) format = value.number.toString();
                        if (value.double) format = value.double.toString();
                        if (value.bool) format = value.bool.toString();
                        if (value.string) format = value.string;
                        if (value.timestamp) format = value.timestamp;
                        return (<Text>{format}</Text>)
                    },
                }))}
                //@ts-ignore
                data={data}
            />}
        </Box>
    )
}


const labels = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];



type ChartType = "line" | "bar" | "axis-mixed" | "pie" | "percentage" | "heatmap";

type AxisMode = "span" | "tick";

type ChartData = {
    labels?: Array<string>;
    datasets?: Array<{
        name?: string;
        chartType?: ChartType;
        values: Array<number>;
    }>;
    dataPoints?: { [x: string]: number };
    start?: Date;
    end?: Date;
};

type SelectEvent = {
    label: string;
    values: number[];
    index: number;
};


type TooltipOptions = {
    formatTooltipX?: (value: number) => any;
    formatTooltipY?: (value: number) => any;
};

type Props = {
    animate?: 0 | 1;
    title?: string;
    type?: ChartType;
    data: ChartData;
    height?: number;
    colors?: Array<string>;
    axisOptions?: {
        xAxisMode?: AxisMode;
        yAxisMode?: AxisMode;
        xIsSeries?: 0 | 1;
    };
    barOptions?: {
        spaceRatio?: number;
        stacked?: 0 | 1;
        height?: number;
        depth?: number;
    };
    lineOptions?: {
        dotSize?: number;
        hideLine?: 0 | 1;
        hideDots?: 0 | 1;
        heatline?: 0 | 1;
        regionFill?: number;
        areaFill?: number;
        spline?: 0 | 1;
    };
    isNavigable?: boolean;
    maxSlices?: number;
    truncateLegends?: 0 | 1;
    tooltipOptions?: TooltipOptions;
    onDataSelect?: (event: SelectEvent) => void;
    valuesOverPoints?: 0 | 1;
};

const FChart = forwardRef((props: Props, parentRef) => {
    const ref = useRef<HTMLDivElement>(null)
    const chart = useRef<any>(null)
    const initialRender = useRef<boolean>(true)
    const { onDataSelect } = props;

    useImperativeHandle(parentRef, () => ({
        export: () => {
            if (chart && chart.current) {
                chart.current.export();
            }
        },
    }))
    useEffect(() => {
        chart.current = new Chart(ref.current, {
            isNavigable: onDataSelect !== undefined,
            ...props,
        });
        if (onDataSelect) {
            chart.current.parent.addEventListener("data-select", (e: SelectEvent & React.SyntheticEvent) => {
                e.preventDefault();
                onDataSelect(e);
            });
        }
    }, [])
    useEffect(() => {
        if (initialRender.current) {
            initialRender.current = false;
            return;
        }
        chart.current.update(props.data);
    }, [props.data])
    return (<Box ref={ref} />)
})