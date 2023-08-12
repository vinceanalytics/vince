

import { Box, ActionMenu, ActionList, Button } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import {
    TriangleRightIcon,
    ZapIcon,
    CodeIcon,
} from "@primer/octicons-react";

export const Menu = () => {
    return (
        <Box
            pl={2}
            pr={3}
            sx={{ borderBottomWidth: 1, borderBottomStyle: 'solid', borderColor: 'border.default', pb: 1 }}
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
                        <Button
                            variant="primary"
                            leadingIcon={ZapIcon}
                            sx={{ mr: 1 }}
                        >Save</Button>
                    </PageHeader.Actions>
                </PageHeader.TitleArea>
            </PageHeader>
        </Box>
    )
}