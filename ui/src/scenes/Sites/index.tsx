import { Box, Button, TreeView } from "@primer/react";
import { TableIcon, PlusIcon, GoalIcon, AlertIcon } from "@primer/octicons-react";

import { UnderlineNav } from '@primer/react/drafts'
import SitePanel from "./sitePanel";
import GoalsPanel from "./goalsPanel";
import AlertsPanel from "./alertsPanel";
import { useState } from "react";



const Sites = () => {
    const [active, setActive] = useState<string>("sites")
    return (
        <Box
            sx={{
                display: "flex",
                overflow: "auto",
                flex: "1",
                flexDirection: "column",
                paddingLeft: "2px",
            }}>
            <UnderlineNav aria-label="Sites" >
                <UnderlineNav.Item
                    aria-current={active === "sites" ? true : undefined}
                    onSelect={() => setActive("sites")}
                    icon={TableIcon}
                >
                    Sites
                </UnderlineNav.Item>
                <UnderlineNav.Item
                    aria-current={active === "goals" ? true : undefined}
                    onSelect={() => setActive("goals")}
                    icon={GoalIcon}
                >
                    Goals
                </UnderlineNav.Item>
                <UnderlineNav.Item
                    aria-current={active === "alerts" ? true : undefined}
                    onSelect={() => setActive("alerts")}
                    icon={AlertIcon}
                >
                    Alerts
                </UnderlineNav.Item>
            </UnderlineNav>
            <Box>
                {active === "sites" && <SitePanel />}
                {active === "goals" && <GoalsPanel />}
                {active === "alerts" && <AlertsPanel />}
            </Box>
        </Box>
    )
}

export default Sites;