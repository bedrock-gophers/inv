# Inv
The Inv library allows the use of inventory menus, providing tools to create or send fake inventories.

## Creating an Inventory Menu
To create an inventory menu using the Inv library, follow these steps:

Handle Player Packets: inv requires you to use the RedirectPlayerPackets on player join which makes it possible
for the library to handle incoming player packets, mostly use to handle container closing.
```go
// The 'conf' variable represents the server config.
conf.Listeners = intercept.WrapListeners(conf.Listeners)
```

Create Menu Submittable: Your menu requires a menu submittable, here's an example:
```go
type MySubmittable struct{}

func (m MySubmittable) Submit(p *player.Player, it item.Stack) {
	fmt.Println("Submitted", it)
}

func (m MySubmittable) Close(p *player.Player) {
	fmt.Println("Closed")
}
```
Create Menu: Use the inv.NewMenu to create a new inventory menu.
```go
m := inv.NewMenu(MySubmittable{}, "test", inv.ContainerTypeChest)
```
Populate Menu Slots: Provide the menu with a `item.Stack` slice:
```go
var stacks = make([]item.Stack, 27)
for i := 0; i < 27; i++ {
    stacks[i] = item.NewStack(block.StainedGlass{Colour: item.ColourRed()}, 1)
}
m = m.WithStacks(stacks...)
```
Sending the Menu to a Player
To display the inventory menu to a player, use the inv.SendMenu function. Here's an example:

```go
// The 'p' variable represents the targeted player.
inv.SendMenu(p, m)
```
This code sends the inventory menu to the specified player.

## Creating a Form Inventory Menu
Form inventories use Bedrock menu forms rendered as a chest by the bundled resource pack. Add the pack to your Dragonfly server config before creating the server:

```go
import "github.com/bedrock-gophers/inv/forminv"

if err := forminv.AddToConfig(&conf); err != nil {
    log.Fatalln(err)
}
conf.ResourcesRequired = true
```

Register the bundled pack once. It overrides Bedrock server-form UI definitions so menus created by `forminv` render as chest grids while preserving normal form submission.

Create and send a form inventory menu with the `forminv` package:

```go
import (
    "github.com/bedrock-gophers/inv/forminv"
    "github.com/df-mc/dragonfly/server/player"
)

type MyFormMenu struct{}

func (MyFormMenu) Submit(p *player.Player, slot forminv.Slot) {
    fmt.Println("Submitted", slot.Button.Text)
}

func (MyFormMenu) Close(p *player.Player) {
    fmt.Println("Closed")
}

m := forminv.NewMenu(MyFormMenu{}, "test", forminv.LargeChest).WithSlots(
    forminv.NewSlot(4, "Nodebuff", "textures/items/diamond_sword", "nodebuff"),
    forminv.NewSlot(22, "Settings", "textures/ui/icon_setting", "settings"),
)
forminv.SendMenu(p, m)
```

Use `NewSlot` or `slot.At(index)` to place a slot at a specific index. Unpositioned slots fill the first open slot, and empty slots are rendered as placeholders. Use `WithSlots` when you want to bind a submitted value that is separate from the rendered button.
