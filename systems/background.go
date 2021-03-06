package systems

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"fmt"
	"image/color"
	"math"
	"sync"
)

// Defining the Map system
type MapSystem struct {
	world      *ecs.World
	vert_lines []GridEntity
	hor_lines  []GridEntity
	ChunkBoxes [][]GridEntity

	LinePrevXOffset int
	LinePrevYOffset int
	BoxPrevXOffset  int
	BoxPrevYOffset  int
}

//Place holders to satisfy Interface

func (*MapSystem) Remove(ecs.BasicEntity) {}

// Every object of this entity is one grid line
type GridEntity struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

var item_tobe_placed int = 0
var mouseheld bool = false

// When system is created this func is executed
// Initialze the world variable and assign tab to toggle the grid
func (ms *MapSystem) New(w *ecs.World) {
	ms.world = w
	GridSize = 32
	ms.LinePrevXOffset = 0
	ms.LinePrevYOffset = 0
	ms.BoxPrevXOffset = 0
	ms.BoxPrevYOffset = 0

	// Initializes the Grid lines
	func() {
		//Calculates how many vertical and horizontal grid lines we need
		vert_num := int(engo.WindowWidth()) / GridSize
		hor_num := int(engo.WindowHeight()) / GridSize

		//Each grid line is an Entity, so we are storing all vert and hor lines in two
		//Seperate slices
		ms.vert_lines = make([]GridEntity, vert_num)
		ms.hor_lines = make([]GridEntity, hor_num)

		//Generating the vert grid lines
		for i := 0; i < vert_num; i++ {
			ms.vert_lines[i] = GridEntity{
				BasicEntity: ecs.NewBasic(),
				RenderComponent: common.RenderComponent{
					Drawable: common.Rectangle{},
					Color:    color.RGBA{0, 0, 0, 125},
				},
				SpaceComponent: common.SpaceComponent{
					Position: engo.Point{float32(i * GridSize), 0},
					Width:    2,
					Height:   engo.WindowHeight(),
				},
			}
			ms.vert_lines[i].RenderComponent.SetZIndex(80)
			ms.vert_lines[i].RenderComponent.SetShader(common.HUDShader)
		}
		//Generating the hor grid lines
		for i := 0; i < hor_num; i++ {
			ms.hor_lines[i] = GridEntity{
				BasicEntity: ecs.NewBasic(),
				RenderComponent: common.RenderComponent{
					Drawable: common.Rectangle{},
					Color:    color.RGBA{0, 0, 0, 125},
				},
				SpaceComponent: common.SpaceComponent{
					Position: engo.Point{0, float32(i * GridSize)},
					Width:    engo.WindowWidth(),
					Height:   2,
				},
			}
			// Make the grid HUD, at a depth between 0 and HUD's
			ms.hor_lines[i].RenderComponent.SetZIndex(80)
			ms.hor_lines[i].RenderComponent.SetShader(common.HUDShader)
		}

		// Add each grid line entity to the render system
		for i := 0; i < vert_num; i++ {
			ActiveSystems.RenderSys.Add(&ms.vert_lines[i].BasicEntity, &ms.vert_lines[i].RenderComponent, &ms.vert_lines[i].SpaceComponent)
			ms.vert_lines[i].RenderComponent.Hidden = true
		}
		for i := 0; i < hor_num; i++ {
			ActiveSystems.RenderSys.Add(&ms.hor_lines[i].BasicEntity, &ms.hor_lines[i].RenderComponent, &ms.hor_lines[i].SpaceComponent)
			ms.hor_lines[i].RenderComponent.Hidden = true
		}
	}()

	// Initializes the Chunk Rectangles
	func() {
		per_row := int(math.Ceil(float64(engo.WindowWidth())/float64(GridSize*ChunkSize))) + 1
		per_col := int(engo.WindowHeight())/(GridSize*ChunkSize) + 1

		ms.ChunkBoxes = make([][]GridEntity, per_row)
		for i := 0; i < per_row; i++ {
			ms.ChunkBoxes[i] = make([]GridEntity, per_col)
		}

		for i := 0; i < per_row; i++ {
			for j := 0; j < per_col; j++ {
				ms.ChunkBoxes[i][j] = GridEntity{
					BasicEntity: ecs.NewBasic(),
					SpaceComponent: common.SpaceComponent{
						Position: engo.Point{float32(i * GridSize * ChunkSize), float32(j * GridSize * ChunkSize)},
						Width:    float32(GridSize * ChunkSize),
						Height:   float32(GridSize * ChunkSize),
					},
					RenderComponent: common.RenderComponent{
						Drawable: common.Rectangle{
							BorderWidth: 2,
							BorderColor: color.RGBA{255, 255, 255, 255},
						},
						Color: color.RGBA{0, 0, 0, 0},
					},
				}

				ms.ChunkBoxes[i][j].SetShader(common.HUDShader)
				ms.ChunkBoxes[i][j].SetZIndex(81)
				ms.ChunkBoxes[i][j].Hidden = true

				cb := &ms.ChunkBoxes[i][j]
				ActiveSystems.RenderSys.Add(&cb.BasicEntity, &cb.RenderComponent, &cb.SpaceComponent)
			}
		}
	}()

	fmt.Println("Map System initialized")
}

func (ms *MapSystem) Update(dt float32) {

	//mx, my := GetAdjustedMousePos(false)

	// // Map editing code
	// func() {
	// 	if engo.Input.Mouse.Action == engo.Press && engo.Input.Mouse.Button == engo.MouseButtonRight {
	// 		fmt.Println(item_tobe_placed)
	// 		item_tobe_placed += 1
	// 		var BuildingName string

	// 		switch item_tobe_placed % 6 {
	// 		case 0:
	// 			BuildingName = "Tree"
	// 		case 1:
	// 			BuildingName = "Bush"
	// 		case 2:
	// 			BuildingName = "House"
	// 		case 3:
	// 			BuildingName = "Town Center"
	// 		case 4:
	// 			BuildingName = "Military Block"
	// 		case 5:
	// 			BuildingName = "Resource Building"
	// 		default:
	// 			panic("Math is broken!")
	// 		}

	// 		fmt.Println("Left Click Now places", BuildingName)
	// 	}
	// 	if engo.Input.Mouse.Action == engo.Press && engo.Input.Mouse.Button == engo.MouseButtonLeft {
	// 		mouseheld = true
	// 	}
	// 	if engo.Input.Mouse.Action == engo.Release && engo.Input.Mouse.Button == engo.MouseButtonLeft {
	// 		mouseheld = false
	// 	}
	// 	if mouseheld {
	// 		var BuildingName string

	// 		switch item_tobe_placed % 6 {
	// 		case 0:
	// 			BuildingName = "Tree"
	// 		case 1:
	// 			BuildingName = "Bush"
	// 		case 2:
	// 			BuildingName = "House"
	// 		case 3:
	// 			BuildingName = "Town Center"
	// 		case 4:
	// 			BuildingName = "Military Block"
	// 		case 5:
	// 			BuildingName = "Resource Building"
	// 		default:
	// 			panic("Math is broken!")
	// 		}
	// 		if WithinGameWindow(mx, my) {
	// 			pik := float32(math.Floor(float64(mx)/float64(GridSize)) * float64(GridSize))
	// 			cik := float32(math.Floor(float64(my)/float64(GridSize)) * float64(GridSize))
	// 			pik = float32(math.Floor(float64(pik / float32(GridSize))))
	// 			cik = float32(math.Floor(float64(cik / float32(GridSize))))
	// 			if !Grid[int(pik)][int(cik)] {
	// 				engo.Mailbox.Dispatch(CreateBuildingMessage{Name: BuildingName, Position: engo.Point{X: pik * float32(GridSize), Y: cik * float32(GridSize)}})
	// 				fmt.Println("Create")
	// 			} else {
	// 			}
	// 		}
	// 	}
	// 	if engo.Input.Button(R_remove).Down() {
	// 		se := GetStaticHover()
	// 		if se != nil {
	// 			engo.Mailbox.Dispatch(DestroyBuildingMessage{obj: GetStaticHover()})
	// 		}
	// 	}

	// 	if engo.Input.Button(SaveKey).JustReleased() {
	// 		engo.Mailbox.Dispatch(SaveMapMessage{Fname: "World.mapfile"})
	// 	}
	// }()

	//Rendering the Gridlines and Chunk Boxes
	func() {
		// Toggle the hidden attribute of every grid line's render component
		if engo.Input.Button(GridToggle).JustPressed() {
			for i, _ := range ms.vert_lines {
				ms.vert_lines[i].RenderComponent.Hidden = !ms.vert_lines[i].RenderComponent.Hidden
			}
			for i, _ := range ms.hor_lines {
				ms.hor_lines[i].RenderComponent.Hidden = !ms.hor_lines[i].RenderComponent.Hidden
			}
			for i, _ := range ms.ChunkBoxes {
				for j, _ := range ms.ChunkBoxes[i] {
					ms.ChunkBoxes[i][j].RenderComponent.Hidden = !ms.ChunkBoxes[i][j].RenderComponent.Hidden
				}
			}
		}

		if ms.vert_lines[0].RenderComponent.Hidden == false {
			CamSys := ActiveSystems.CameraSys

			LineXOffset := int(CamSys.X()) % GridSize
			LineYOffset := int(CamSys.Y()) % GridSize
			BoxXOffset := int(CamSys.X()-engo.WindowWidth()/2) % (GridSize * ChunkSize)
			BoxYOffset := int(CamSys.Y()-engo.WindowHeight()/2) % (GridSize * ChunkSize)

			wg := sync.WaitGroup{}

			wg.Add(3)
			// Updating hor and vert lines in parallel for faster execution
			go func() {
				defer wg.Done()
				for i, _ := range ms.vert_lines {
					ms.vert_lines[i].Position.Add(engo.Point{float32(ms.LinePrevXOffset-LineXOffset) * CamSys.Z() * (engo.GameWidth() / engo.CanvasWidth()), 0})
				}
			}()

			go func() {
				defer wg.Done()
				for i, _ := range ms.hor_lines {
					ms.hor_lines[i].Position.Add(engo.Point{0, float32(ms.LinePrevYOffset-LineYOffset) * CamSys.Z() * (engo.GameHeight() / engo.CanvasHeight())})
				}
			}()

			go func() {
				defer wg.Done()
				for i, _ := range ms.ChunkBoxes {
					for j, _ := range ms.ChunkBoxes[i] {
						ms.ChunkBoxes[i][j].Position.Add(engo.Point{float32(ms.BoxPrevXOffset-BoxXOffset) * CamSys.Z() * (engo.GameWidth() / engo.CanvasWidth()), float32(ms.BoxPrevYOffset-BoxYOffset) * CamSys.Z() * (engo.GameHeight() / engo.CanvasHeight())})
					}
				}
			}()
			wg.Wait()

			ms.LinePrevXOffset = LineXOffset
			ms.LinePrevYOffset = LineYOffset
			ms.BoxPrevXOffset = BoxXOffset
			ms.BoxPrevYOffset = BoxYOffset
		}
	}()

	// Handle Middle Mouse clicks for debugging
	func() {
		mx, my := GetAdjustedMousePos(false)

		if engo.Input.Mouse.Action == engo.Press && engo.Input.Mouse.Button == engo.MouseButtonMiddle {
			fmt.Println("---------------------------------------------")
			fmt.Println("Mouse Pos is", mx, "(", int(mx)/GridSize, "),", my, "(", int(my)/GridSize, ")")
			ChunkRef, ChunkIndex := GetChunkFromPos(mx, my)
			Chunk := *ChunkRef
			Sector, SectorIndex := GetSectorFromPos(mx, my)

			if len(Chunk) > 0 {
				fmt.Println("-------------------------")
				for _, item := range Chunk {
					fmt.Println(item.GetStaticComponent().Name, "present in chunk:", ChunkIndex)
				}
			} else {
				fmt.Println("Chunk", ChunkIndex, "Empty")
			}
			fmt.Println("-------------------------")
			if len(*Sector) > 0 {
				for _, item := range *Sector {
					fmt.Println(item.Name, "present in chunk:", SectorIndex)
				}
			} else {
				fmt.Println("Sector", ChunkIndex, "Empty")
			}
			fmt.Println("-------------------------")

			if GetGridAtPos(mx, my) {
				fmt.Println("Grid at", int(mx)/GridSize, ",", int(my)/GridSize, "is occupied")
			} else {
				fmt.Println("Grid at", int(mx)/GridSize, ",", int(my)/GridSize, "is not occupied")
			}
			fmt.Println("---------------------------------------------")
		}
	}()
}
