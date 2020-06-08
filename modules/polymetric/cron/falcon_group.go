package cron

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	"github.com/open-falcon/falcon-plus/modules/polymetric/model"
)

func (this *GeneralPoly) FalconGroupwork(name string, strategys []*model.PolyMetric) {

	ends := RunFalconGroupSql(name)
	for _, item := range strategys {
		a := &GroupRes{}
		a.Name = name
		a.Counter = item.Counter
		a.Ends = ends
		this.Result <- a
		//log.Infof("FalconGroupwork_%+v", a)
	}
}

func RunFalconGroupSql(grpName string) (ends []string) {
	Sql := fmt.Sprintf("select hostname  from host a ,grp_host b,grp c where a.id=b.host_id and b.grp_id=c.id and c.grp_name='%s'", grpName)

	Q := orm.NewOrm()
	_, error := Q.Raw(Sql).QueryRows(&ends)
	if error != nil {
		log.Errorf("RunFalconGroupSql:Query_ends_error:%s,%+v", grpName, error)
		return
	}
	return

}

func RenewFalconGroupStrategy() {
	res, num := CommonInitQueue(FalconGroupPolyType)
	if res == nil {
		return
	}
	gp := &GeneralPoly{}
	gp.Num = num
	gp.ActionFunc = gp.FalconGroupwork
	gp.ArgMap = res
	gp.Type = FalconGroupPolyType
	MultiRunWorker(gp)
}
