import { ChevronDownIcon } from "@primer/octicons-react"
import { PageLayout, Box, Heading, Button, Octicon, ActionMenu, ActionList, Text, TokenProps, TextInputWithTokens, Select, FormControl, TextInput } from "@primer/react"
import { useCallback, useState } from "react"
import { Dialog, DataTable } from '@primer/react/drafts'



export const Dashboard = () => {
    const [sites, setSites] = useState<string[]>(["vinceanalytics.com", "example.com"])
    const [active, setActive] = useState<number>(0)
    const [tokens, setTokens] = useState<TokenProps[]>([
        { id: 0, text: "path==/" },
    ])

    const onTokenRemove = (tokenId: string | number) => {
        setTokens(tokens.filter(token => token.id !== tokenId))
    }

    const onTokenAdd = (tokenId: string) => {
        setTokens([{ id: tokens.length, text: tokenId }, ...tokens])
    }

    return (
        <PageLayout>
            <PageLayout.Header>
                <Box py={2} zIndex={10}>
                    <Box display={"flex"} width="100%" alignItems={"center"}>
                        <SitesSelection active={active} sites={sites} setActive={setActive} />
                        <Filter tokens={tokens} onAdd={onTokenAdd} onRemove={onTokenRemove} />
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
                            return undefined
                        }
                        return <ActionList.Item key={value} onClick={() => setActive(idx)}><Text sx={{ fontWeight: "bold" }}>{value}</Text></ActionList.Item>
                    })}
                </ActionList>
            </ActionMenu.Overlay>
        </ActionMenu>
    )
}


type FilterProps = {
    tokens: TokenProps[]
    onRemove: (tokenId: string | number) => void
    onAdd: (tokenId: string) => void
}

const Filter = ({ tokens, onRemove, onAdd }: FilterProps) => {
    const [isOpen, setOpen] = useState<boolean>(false)
    const openDialog = useCallback(() => setOpen(true), [setOpen])
    const closeDialog = useCallback(() => setOpen(false), [setOpen])

    return (
        <Box>
            {isOpen && (
                <Dialog
                    title="Add filter"
                    onClose={closeDialog}
                >
                    <FilterItems onAdd={(filter) => {
                        onAdd(filter)
                        closeDialog()
                    }} />
                </Dialog>
            )}
            <TextInputWithTokens placeholder="filters" tokens={tokens} onTokenRemove={onRemove}
                onClick={openDialog}
            />
        </Box>
    )
}





const properties = [
    "event",
    "page",
    "entry_page",
    "exit_page",
    "source",
    "referrer",
    "utm_source",
    "utm_medium",
    "utm_campaign",
    "utm_content",
    "utm_term",
    "device",
    "browser",
    "browser_version",
    "os",
    "os_version",
    "country",
    "region",
    "domain",
    "city"
]

interface Op {
    name: string
    value: string
}

const ops: Op[] = [
    { name: "equal", value: "==" },
    { name: "not_equal", value: "!=" },
    { name: "regex_equal", value: "~=" },
    { name: "regex_not_equal", value: "!~" }
]



type FilterItemsProps = {
    onAdd: (filter: string) => void
}

const FilterItems = ({ onAdd }: FilterItemsProps) => {
    const [property, setProperty] = useState<string>()
    const [op, setOp] = useState<string>()
    const [value, steValue] = useState<string>()
    const submit = useCallback(() => {
        if (property && op && value) {
            onAdd(property + op + value)
        }
    }, [property, op, value, onAdd])
    return (
        <Box display={"flex"}>
            <Select placeholder="Select Property" onChange={(e) => {
                setProperty(e.target.value)
            }}>
                {properties.map((value) =>
                    <Select.Option key={value} value={value}>{value}</Select.Option>
                )}
            </Select>
            <Select placeholder="Select Operation" onChange={(e) => {
                setOp(e.target.value)
            }}>
                {ops.map((value) =>
                    <Select.Option key={value.value} value={value.value}>{value.name}</Select.Option>
                )}
            </Select>
            <TextInput placeholder="Value" onChange={(e) => {
                steValue(e.target.value)
            }} required>
            </TextInput>
            <Button variant="primary" sx={{ px: 5, mx: 2 }} onClick={submit} >Add</Button>
        </Box>
    )
}