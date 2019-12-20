package models

import (
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/jinzhu/gorm"
	"github.com/youlu-cn/ginp"
)

const (
	TopicStatusEnable = iota
	TopicStatusDisable
)

type Topic struct {
	ginp.Model

	System         string `binding:"required" json:"system"`
	Queue          string `binding:"required" json:"queue" gorm:"unique_index:uix_name_queue_name"`
	Name           string `binding:"required" json:"name" gorm:"unique_index:uix_name_queue_name"`
	Describe       string `binding:"required" json:"desc" gorm:"column:description"`
	ConsumeFile    string `binding:"required" json:"consume"`
	RetryTimes     int    `json:"retry_times" gorm:"default:'3'"`
	MaxQueueLength int    `json:"max_queue_length" gorm:"default:'1000'"`
	RunType        int    `json:"run_type" gorm:"default:'0'"`
	Password       string `json:"password"`
	NumOfWorkers   int    `json:"num_of_workers" gorm:"default:'10'"`
	CgiConfig      string `json:"cgi_config" gorm:"default:'local'"`
	Storage        uint64 `binding:"required" json:"storage"`
	Status         uint8  `json:"status" gorm:"default:0"`
	Alarm          int    `json:"alarm" gorm:"default:0"`
	AlarmRetry     int    `json:"alarm_retry" gorm:"default:0"`
	HttpConfig     string `json:"http_config" gorm:"default:''"`
	TopicConfig    string `json:"topic_config" gorm:"default:''"`

	// operator
	Comment  string `binding:"required" json:"comment"`
	Operator string `json:"author"`
}

func (m Topic) TableName() string {
	return "topic"
}

func (m Topic) Create() (err error) {

	tx := dao.DB.Begin()
	if err = tx.Error; err != nil {
		return
	}
	if err = tx.Create(&m).Error; err != nil {
		goto Rollback
	}
	if err = tx.Model(&System{}).Where("name = ?", m.System).
		Update("topic_count", gorm.Expr("topic_count + 1")).Error; err != nil {
		goto Rollback
	}
	if err = tx.Model(&Queue{}).Where("name = ?", m.Queue).
		Update("topic_count", gorm.Expr("topic_count + 1")).Error; err != nil {
		goto Rollback
	}
	return tx.Commit().Error

Rollback:
	if tx.Rollback().Error != nil {
		// todo: add log
	}
	return
}

func (m Topic) Delete() (err error) {
	tx := dao.DB.Begin()
	if err = tx.Error; err != nil {
		return
	}
	if err = tx.Delete(&m).Error; err != nil {
		goto Rollback
	}
	if err = tx.Model(&System{}).Where("name = ?", m.System).
		Update("topic_count", gorm.Expr("topic_count - 1")).Error; err != nil {
		goto Rollback
	}
	if err = tx.Model(&Queue{}).Where("name = ?", m.Queue).
		Update("topic_count", gorm.Expr("topic_count - 1")).Error; err != nil {
		goto Rollback
	}
	return tx.Commit().Error

Rollback:
	if tx.Rollback().Error != nil {
		// todo: add log
	}
	return
}

func (m Topic) Update() error {
	return dao.DB.Save(&m).Error
}

func (m Topic) Get(id uint64) (v *Topic, err error) {
	db := dao.DB.Model(&m).Where("id = ?", id).First(&m)
	return &m, db.Error
}

func (m Topic) Query(query *Query) (v []*Topic, err error) {
	db := dao.DB.Model(&m)
	if query.System != "" {
		db = db.Where("system = ?", query.System)
	}
	if query.Queue != "" {
		db = db.Where("queue = ?", query.Queue)
	}
	db = db.Find(&v)

	return v, db.Error
}

func (m Topic) Search(system, queue, topic string, con string) (v []*Topic, err error) {
	db := dao.DB.Model(&m).Where("(system like ? and queue like ?) "+con+" name like ?", "%"+system+"%", queue, topic).Limit(100)
	db = db.Find(&v)
	return v, db.Error
}

func (m *Topic) FindOne(query map[string]string) error {
	db := dao.DB.Model(&m)
	for _, k := range []string{"queue", "name"} {
		if v, ok := query[k]; ok {
			db = db.Where(k+" = ?", v)
		}
	}
	db = db.First(&m)
	return db.Error
}

func (m Topic) FindAllByQueues(queueNames []string) ([]TransModelTopic, error) {
	var v []TransModelTopic
	sql := `select t.system, t.queue, t.name, t.description, t.password, t.consume_file, t.comment, t.status, s.type,s.host,s.port
			from topic as t left join storage as s on t.storage=s.id where queue in (?)`
	rows, err := dao.DB.Raw(sql, queueNames).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		system, queue, name, desc, password, consume, comment, status, typeVal, host, port string
	)
	for rows.Next() {
		_ = rows.Scan(&system, &queue, &name, &desc, &password, &consume, &comment, &status, &typeVal, &host, &port)
		v = append(v, TransModelTopic{
			System:   system,
			Queue:    queue,
			Name:     name,
			Desc:     desc,
			Password: password,
			Consume:  consume,
			Comment:  comment,
			Status:   status,
			Storage: TransModelStorage{
				SType: typeVal,
				Host:  host,
				Port:  port,
			},
		})
	}
	return v, nil
}
