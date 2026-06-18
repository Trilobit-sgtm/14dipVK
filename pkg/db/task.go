package db

import "fmt"

// Task описывает одну запись в таблице scheduler.
// Все поля — строки, чтобы удобно сериализовать/десериализовать JSON.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в таблицу и возвращает id новой записи.
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat)
	          VALUES (?, ?, ?, ?)`

	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return id, err
}

// Tasks возвращает ближайшие limit задач, отсортированных по дате.
func Tasks(limit int) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat
	          FROM scheduler
	          ORDER BY date
	          LIMIT ?`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// создаём пустой слайс, чтобы JSON вернул [] а не null
	tasks := make([]*Task, 0)

	for rows.Next() {
		t := &Task{}
		if err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// GetTask возвращает задачу по идентификатору.
func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat
	          FROM scheduler WHERE id = ?`

	t := &Task{}
	err := db.QueryRow(query, id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	return t, err
}

// UpdateTask обновляет все поля задачи по её id.
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler
	          SET date = ?, title = ?, comment = ?, repeat = ?
	          WHERE id = ?`

	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача с id=%s не найдена", task.ID)
	}
	return nil
}

// UpdateDate меняет только дату задачи — используется при отметке выполнения.
func UpdateDate(id, next string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`

	res, err := db.Exec(query, next, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача с id=%s не найдена", id)
	}
	return nil
}

// DeleteTask удаляет задачу по идентификатору.
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`

	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача с id=%s не найдена", id)
	}
	return nil
}
