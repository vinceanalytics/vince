import { Box } from "@primer/react";
import { TableIcon } from "@primer/octicons-react";

import { UnderlineNav } from '@primer/react/drafts'






const Sites = () => {
    return (
        <Box
            sx={{
                display: "flex",
                overflow: "auto",
                flex: "1",
                flexDirection: "column",
                paddingLeft: "2px",
            }}>
            <UnderlineNav aria-label="Sites">
                <UnderlineNav.Item
                    aria-current="page"
                    icon={TableIcon}
                >Sites</UnderlineNav.Item>
            </UnderlineNav>
        </Box>
    )
}

export default Sites;