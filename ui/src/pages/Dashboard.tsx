import { ChevronDownIcon } from "@primer/octicons-react"
import { PageLayout, Box, Heading, Button, Octicon, ActionMenu, ActionList, Text } from "@primer/react"
import { useState } from "react"

export const Dashboard = () => {
    return (
        <PageLayout>
            <PageLayout.Header>
                <Box py={2} zIndex={10}>
                    <Box display={"flex"} width="100%" alignItems={"center"}>
                        <SitesSelection active={0} sites={["vinceanalytics.com", "example.com"]} />
                    </Box>
                </Box>
            </PageLayout.Header>
            <PageLayout.Content>
                <Box m={4}>
                    <Heading as="h2" sx={{ mb: 2 }}>
                        Hello, world!
                    </Heading>
                    <p>This will get Primer text styles.</p>
                </Box>
            </PageLayout.Content>
        </PageLayout>
    )
}


type SiteSelectionProps = {
    active: number
    sites: string[]
}

const SitesSelection = (props: SiteSelectionProps) => {
    const [active, setActive] = useState<number>(props.active)
    return (
        <ActionMenu>
            <ActionMenu.Button><Text sx={{ fontWeight: "bold" }}>{props.sites[active]}</Text></ActionMenu.Button>
            <ActionMenu.Overlay>
                <ActionList>
                    {props.sites.map((value, idx) => {
                        if (idx == active) {
                            return <div></div>
                        }
                        return <ActionList.Item onClick={() => setActive(idx)}><Text sx={{ fontWeight: "bold" }}>{value}</Text></ActionList.Item>
                    })}
                </ActionList>
            </ActionMenu.Overlay>
        </ActionMenu>
    )
}