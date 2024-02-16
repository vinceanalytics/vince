import { ActionList, ActionMenu, Box, Text } from "@primer/react"
import { Site } from "../vince"


type SiteProps = {
    active: string,
    sites: Site[]
    selectSite: (name: string) => void,
}

export const SitesSelector = ({ active, sites, selectSite }: SiteProps) => {

    return (
        <ActionMenu>
            <ActionMenu.Button><Text sx={{ fontWeight: "bold" }}>{active}</Text></ActionMenu.Button>
            <ActionMenu.Overlay>
                <ActionList>
                    {sites.filter((e) => e.name !== active).map(({ name }, idx) => {
                        return <ActionList.Item key={name} onClick={() => selectSite(name)}><Text sx={{ fontWeight: "bold" }}>{name}</Text></ActionList.Item>
                    })}
                </ActionList>
            </ActionMenu.Overlay>
        </ActionMenu>
    )
}



