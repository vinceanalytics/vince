import { Box, NavList } from "@primer/react"
import { SettingsContextPros } from "./types"
import { useSites } from "../../providers"
import { useState } from "react"

export const SitesSettingsContext = ({ item }: SettingsContextPros) => {
    const { sites } = useSites()
    const [selectedSite, setSelectedSite] = useState<string>()
    return (
        <Box
            display={item === "sites" ? "grid" : "none"}
            gridTemplateColumns={"auto"}
        >
            <NavList>
                {sites.map((site) => (
                    <NavList.Item
                        aria-current={selectedSite === site.domain}
                        onClick={() => setSelectedSite(site.domain)}
                    >
                        {site.domain}
                    </NavList.Item>
                ))}
            </NavList>

        </Box>
    )
}