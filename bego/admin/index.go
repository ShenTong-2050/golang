package admin

import (
	"bego/models/adminmodel"
	"fmt"
	"github.com/beego/beego/v2/core/utils"
	"github.com/beego/beego/v2/core/validation"
	"github.com/beego/beego/v2/server/web"
	"html/template"
	"strconv"
	"strings"
)

type IndexController struct {
	baseController
}

type IndexData struct {
	MarkOnline                string
	Gamedata                  []adminmodel.GameInfo
	Gamelist                  []adminmodel.AppProvider
	Selname                   string
	Quanbu, Online, Underline bool
	Active, Open              int
}

// 后台首页
func (this *IndexController) Index() {
	var gameList, fenfa, datainfo, gamedata []adminmodel.GameInfo
	var num int64
	var index IndexData
	index.Active = 0
	index.Open = 0

	// 接收参数
	per, _ := web.AppConfig.Int("page::perPage")
	p, _ := utils.StrTo(this.Ctx.Input.Query("p")).Int()
	mark_online := this.Ctx.Input.Query("mark_online")
	app_name := this.Ctx.Input.Query("app_name")

	if p > 0 {
		p = (p - 1) * per
	}

	// 获取全部分发游戏
	gameList, num = adminmodel.FindAll(p, per)
	alreadyGid := []string{}
	for i := range gameList {
		alreadyGid = append(alreadyGid, gameList[i].Gid)
	}

	// 可创建分发游戏列表
	var listenFirst bool
	if this.GetSession("listenfirst") == nil {
		listenFirst = false
	} else {
		listenFirst = true
	}
	gamemsg := adminmodel.GetGid(listenFirst, strings.Join(alreadyGid, ","))

	index.Gamelist = gamemsg

	// 获取全部游戏数据
	if mark_online == "" && app_name == "" {
		// 获取全部创建分发的游戏与分页
		//fenfa := admin.FindAll(p,per)
		fenfa, num = adminmodel.FindAll(p, per)
		index.Gamedata = fenfa
	} else {
		if mark_online != "" {
			imk, _ := strconv.Atoi(mark_online)
			// 验证 mark_online 参数是否合法
			valid := validation.Validation{}
			valid.Required(mark_online, "mark_online").Message("非法操作!")
			valid.Range(imk, 0, 2, "mark_online").Message("参数异常!")

			if valid.HasErrors() {
				for _, err := range valid.Errors {
					this.Data["json"] = ReturnData{
						Code: 400,
						Data: nil,
						Info: err.Message,
					}
					err := this.ServeJSON()
					if err != nil {
						return
					}
				}
			}
			if mark_online == "2" {
				// 获取全部游戏
				//datainfo := admin.FindAll(p,per)
				datainfo, num = adminmodel.FindAll(p, per)
				index.Gamedata = datainfo
			} else {
				// 获取上线/下线游戏
				//datainfo := admin.FindGame("mark_online", mark_online,p,per)
				datainfo, num = adminmodel.FindGame("mark_online", mark_online, p, per)
				index.Gamedata = datainfo
			}
			index.MarkOnline = mark_online
		}

		if app_name != "" {
			// 验证游戏名称
			valid := validation.Validation{}
			valid.Required(app_name, "app_name").Message("请求参数异常!")
			valid.Alpha(app_name, "app_name").Message("请输入游戏拼音!")
			if valid.HasErrors() {
				for _, err := range valid.Errors {
					this.Data["json"] = ReturnData{
						Code: 400,
						Data: nil,
						Info: err.Message,
					}
					err := this.ServeJSON()
					if err != nil {
						return
					}
				}
			}
			//gamedata := admin.FindGame("app_name", app_name,p,per)
			gamedata, num = adminmodel.FindGame("app_name", app_name, p, per)
			index.Gamedata = gamedata
			index.Selname = app_name
		}
	}

	// 模板逻辑
	if mark_online == "2" || mark_online == "" {
		index.Quanbu = true
	}

	if mark_online == "1" {
		index.Online = true
	}

	if mark_online == "0" {
		index.Underline = true
	}

	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())

	// 前台显示分页
	this.Data["paginator"] = this.SetPaginator(per, num)
	this.Data["Index"] = index
	this.TplName = "admin/index/index.html"
}

func (this *IndexController) Gamedel() {
	gid := this.Ctx.Input.Query("gid")
	if gid == "" {
		this.Data["json"] = ReturnData{
			Code: 402,
			Data: nil,
			Info: "参数异常...",
		}
		err := this.ServeJSON()
		if err != nil {
			return
		}
	}
	num := adminmodel.Gamedel(gid)
	if num == 0 {
		this.Data["json"] = ReturnData{
			Code: 500,
			Data: nil,
			Info: "删除失败...",
		}
		err := this.ServeJSON()
		if err != nil {
			return
		}
	}
	this.Data["json"] = ReturnData{
		Code: 200,
		Data: nil,
		Info: "删除成功!",
	}
	err := this.ServeJSON()
	if err != nil {
		return
	}
}

func (this *IndexController) Gamesel() {
	this.EnableXSRF = false
	gamelist := adminmodel.GameInfoSelect()
	this.Data["json"] = ReturnData{
		Code: 200,
		Data: gamelist,
		Info: "请求成功!",
	}
	err := this.ServeJSON()
	if err != nil {
		return
	}
}

// 监听用户是否为第一次登录(游戏下拉)
func (this *IndexController) Listen() {
	this.EnableXSRF = false
	data := this.Ctx.Input.Query("isfirst")
	fmt.Println("this is first?", data)
	if data != "" {
		err := this.SetSession("listenfirst", data)
		if err != nil {
			return
		}
		this.Data["json"] = ReturnData{
			Code: 200,
			Data: nil,
			Info: "success",
		}
		err = this.ServeJSON()
		if err != nil {
			return
		}
	} else {
		this.Data["json"] = ReturnData{
			Code: -1,
			Data: nil,
			Info: "Unexpected error",
		}
		err := this.ServeJSON()
		if err != nil {
			return
		}
	}
}

func (this *IndexController) MarkOnline() {
	gid := this.Ctx.Input.Query("gid")
	mark_online := this.Ctx.Input.Query("mark_online")
	imk, _ := strconv.Atoi(mark_online)

	// 验证参数
	valid := validation.Validation{}
	valid.Required(gid, "gid").Message("参数异常...")
	valid.Length(gid, 16, "gid").Message("参数异常...")
	valid.Range(imk, 0, 2, "mark_online").Message("参数异常...")

	if mark_online == "1" {

	}
}
