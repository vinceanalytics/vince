

import { Box, Button, LabelGroup, Label, Spinner } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import {
    TriangleRightIcon, XIcon,
    ZapIcon,
} from "@primer/octicons-react";
import { useState } from "react";
import { useQuery } from "../../../providers";

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
                        <Run />
                    </PageHeader.Title>
                    <PageHeader.Actions>
                        <LabelGroup visibleChildCount={5}>
                            <Label sx={{ cursor: "pointer" }}>page_views</Label>
                            <Label>unique_visitors</Label>
                            <Label>bounce_rate</Label>
                            <Label>visits</Label>
                            <Label>visit_duration</Label>
                            <Label>custom_label</Label>
                        </LabelGroup>
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


const Run = () => {
    const { running, toggleRun } = useQuery()
    return (
        <>
            {running && <Spinner />}
            <Button
                leadingIcon={running ? XIcon : TriangleRightIcon}
                onClick={toggleRun}
                variant={running ? "danger" : "outline"}
            >
                {running ? "Cancel" : "Run"}
            </Button>
        </>
    )
}