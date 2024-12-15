You might like to generate code based on events from a contract.

Let's say you have the ERC-20 ABI:

```json
[
    {
        "name": "Transfer",
        "type": "event",
        "inputs": [
            {"name": "from", "type": "address"},
            {"name": "to", "type": "address"},
            {"name": "value", "type": "uint256"}
        ]
    }
]
```

```go
type Transfer struct {
    Owner   common.Address `json:"owner"`
    Spender common.Address `json:"spender"`
    Value   *big.Int       `json:"value"`
}
```
