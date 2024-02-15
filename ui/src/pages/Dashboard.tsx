import { ChevronDownIcon } from "@primer/octicons-react"
import { PageLayout, Box, Heading, Button, Octicon, ActionMenu, ActionList, Text } from "@primer/react"
import { useState } from "react"


export const Dashboard = () => {
    const [sites, setSites] = useState<string[]>(["vinceanalytics.com", "example.com"])
    const [active, setActive] = useState<number>(0)

    return (
        <PageLayout>
            <PageLayout.Header>
                <Box py={2} zIndex={10}>
                    <Box display={"flex"} width="100%" alignItems={"center"}>
                        <SitesSelection active={active} sites={sites} setActive={setActive} />
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
    setActive: (n: number) => void
}

const SitesSelection = ({ active, sites, setActive }: SiteSelectionProps) => {
    return (
        <ActionMenu>
            <ActionMenu.Button><Text sx={{ fontWeight: "bold" }}>{sites[active]}</Text></ActionMenu.Button>
            <ActionMenu.Overlay>
                <ActionList>
                    {sites.map((value, idx) => {
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