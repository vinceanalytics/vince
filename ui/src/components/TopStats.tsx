import { Box, Heading, Text } from "@primer/react"
import { Card, CardContent, CardHeader, CardTitle } from "./Card";
export const TopStats = () => {
    return (
        <Box
            marginY={4}
        >
            <Box
                display={"grid"}
                sx={{ gap: 1 }}
                gridTemplateColumns={"repeat(5, minmax(0, 1fr))"}
            >
                <Card>
                    <CardHeader>
                        <CardTitle sx={{ fontSize: 3, fontWeight: "semibold" }}> Visitors</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Heading>1.4M</Heading>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle sx={{ fontSize: 3, fontWeight: "semibold" }}> Visits</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Heading>1.4M</Heading>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle sx={{ fontSize: 3, fontWeight: "semibold" }}> Page views</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Heading>1.4M</Heading>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle sx={{ fontSize: 3, fontWeight: "semibold" }}>Bounce Rate</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Heading>1.4M</Heading>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle sx={{ fontSize: 3, fontWeight: "semibold" }}>Visit Duration</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Heading>1.4M</Heading>
                    </CardContent>
                </Card>
            </Box>

        </Box>
    )
}