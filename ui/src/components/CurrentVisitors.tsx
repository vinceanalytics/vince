import { useEffect, useState } from "react"
import { useVince } from "../providers"
import { Box, CounterLabel, Octicon, Text } from "@primer/react"
import { DotFillIcon } from "@primer/octicons-react"

export const CurrentVisitors = () => {
    const { vince } = useVince()
    const [realtime, setRealtime] = useState<number>(0)
    useEffect(() => {
        vince.realtime().then((value) => {
            if (value.visitors) {
                setRealtime(value.visitors)
            }
        }).catch(console.log)
    }, [vince])
    return (
        <Box>
            <Octicon icon={DotFillIcon} fill="green" />
            <CounterLabel>{realtime}</CounterLabel>
            <Text> current visitors</Text>
        </Box>
    )
}