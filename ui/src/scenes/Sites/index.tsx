import { Box, TreeView } from "@primer/react";
import { TableIcon, PlusIcon } from "@primer/octicons-react";

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
            <UnderlineNav aria-label="Sites" >
                <UnderlineNav.Item
                    aria-current="page"
                    icon={TableIcon}
                >Sites</UnderlineNav.Item>
            </UnderlineNav>
            <Box>
                <nav aria-label="Sites Content">
                    <TreeView aria-label="Sites Content">
                        <TreeView.Item id="root">
                            <TreeView.LeadingVisual>
                                <TreeView.DirectoryIcon />
                            </TreeView.LeadingVisual>
                            ..
                            <TreeView.SubTree>
                                <TreeView.Item id="createSite">
                                    <TreeView.LeadingVisual>
                                        <PlusIcon />
                                    </TreeView.LeadingVisual>
                                    Create Site
                                </TreeView.Item>
                            </TreeView.SubTree>
                        </TreeView.Item>
                    </TreeView>
                </nav>
            </Box>
        </Box>
    )
}

export default Sites;