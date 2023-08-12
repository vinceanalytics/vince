

import { Box, ActionMenu, ActionList, Button, Label } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import {
    TriangleRightIcon, TriangleDownIcon,
    CodeIcon,
} from "@primer/octicons-react";

export const Menu = () => {
    return (
        <Box
            paddingX={2} sx={{ borderBottomWidth: 1, borderBottomStyle: 'solid', borderColor: 'border.default', pb: 1 }}
        >
            <PageHeader>
                <PageHeader.TitleArea>
                    <PageHeader.Title>
                        <Button leadingIcon={TriangleRightIcon}
                            variant="outline"
                        >Run</Button>
                    </PageHeader.Title>
                    <PageHeader.Actions>
                        <ActionMenu>
                            <ActionMenu.Button leadingIcon={CodeIcon}
                                variant="outline">
                                Snippets
                            </ActionMenu.Button>
                            <ActionMenu.Overlay>
                                <ActionList>
                                </ActionList>
                            </ActionMenu.Overlay>
                        </ActionMenu>
                    </PageHeader.Actions>
                </PageHeader.TitleArea>
            </PageHeader>
        </Box>
    )
}