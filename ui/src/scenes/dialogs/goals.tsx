import { useCallback, useState } from "react";
import { useVince } from "../../providers";
import { Goal_Type } from "../../vince";
import { Box, FormControl, IconButton, TextInput, ToggleSwitch, Tooltip } from "@primer/react";
import { GoalIcon } from "@primer/octicons-react";
import { Dialog } from '@primer/react/drafts'
import { SelectSite } from "../../components";

export type CreateGoalDialogProps = {
    afterCreate: () => void;
}

export const CreateGoalDialog = ({ afterCreate }: CreateGoalDialogProps) => {
    const { goalsClient } = useVince()
    const [isOpen, setIsOpen] = useState(false);
    const openDialog = useCallback(() => setIsOpen(true), [setIsOpen])
    const closeDialog = useCallback(() => setIsOpen(false), [setIsOpen])
    const [name, setName] = useState<string>("")
    const [domain, setDomain] = useState<string>("")
    const [goalType, setGoalType] = useState<Goal_Type>(Goal_Type.PATH)
    const [value, setValue] = useState<string>("")
    const [custom, setCustom] = useState<boolean>(false)
    const submit = useCallback(() => {
        setIsOpen(false)
        goalsClient?.createGoal({
            name,
            domain,
            type: goalType,
            value,
        }).then(() => {
            afterCreate()
        })
            .catch((e) => {
                console.log(e)
            })
    }, [domain, setIsOpen, name, goalsClient, goalType, value, afterCreate])

    return (
        <>
            <Tooltip aria-label="Create a new Goal" direction="sw">
                <IconButton aria-label="add-goal" onClick={openDialog} icon={GoalIcon} />
            </Tooltip>
            {isOpen && <Dialog
                title="New Goal"
                footerButtons={
                    [{
                        content: 'Create', onClick: submit,
                    }]
                }
                onClose={closeDialog}
            >
                <Box>
                    <FormControl>
                        <FormControl.Label id="goal-name">Name</FormControl.Label>
                        <TextInput block onChange={e => setName(e.currentTarget.value)} />
                    </FormControl>
                    <FormControl>
                        <FormControl.Label>Select Site</FormControl.Label>
                        <SelectSite
                            selectSite={e => setDomain(e)}
                        />
                    </FormControl>
                    <FormControl>
                        <FormControl.Label id="goal-type">Custom Event</FormControl.Label>
                        <ToggleSwitch
                            onChange={on => setGoalType((on ? Goal_Type.EVENT : Goal_Type.PATH))}
                            checked={custom}
                            onClick={() => setCustom(!custom)}
                            aria-labelledby="goal-type"
                        />
                    </FormControl>
                    <FormControl>
                        <FormControl.Label id="goal-type">Value</FormControl.Label>
                        <TextInput block onChange={e => setValue(e.currentTarget.value)} />
                    </FormControl>
                </Box>
            </Dialog>}
        </>
    )
}
