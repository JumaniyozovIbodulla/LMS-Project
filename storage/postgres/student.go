package postgres

import (
	"backend_course/lms/api/models"
	"backend_course/lms/pkg"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type studentRepo struct {
	db *pgxpool.Pool
}

func NewStudent(db *pgxpool.Pool) studentRepo {
	return studentRepo{
		db: db,
	}
}

func (s *studentRepo) Create(ctx context.Context, student models.AddStudent) (string, error) {

	id := uuid.New()
	fmt.Println(student)
	query := `INSERT INTO
					students (id, first_name, last_name, age, external_id, phone, mail, is_active, password) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	_, err := s.db.Exec(ctx, query, id, student.FirstName, student.LastName, student.Age, student.ExternalId, student.Phone, student.Email, student.IsActive, student.Password)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}


func (s *studentRepo) Update(ctx context.Context, student models.Student) (string, error) {
	query := `
	UPDATE
		students
	SET
		first_name = $2, last_name = $3, age = $4, external_id = $5, phone = $6, mail = $7, updated_at = NOW()
	WHERE 
		id = $1; `

	_, err := s.db.Exec(ctx, query, student.Id, student.FirstName, student.LastName, student.Age, student.ExternalId, student.Phone, student.Email)
	if err != nil {
		return "", err
	}
	return student.Id, nil
}

func (s *studentRepo) UpdateStatus(ctx context.Context, student models.Student) (string, error) {
	fmt.Println(student.IsActive)
	if student.IsActive {
		student.IsActive = false
	} else {
		student.IsActive = true
	}
	fmt.Println(student.IsActive)
	query := `
	UPDATE
		students
	SET
	is_active = $2
	WHERE 
		id = $1;`

	_, err := s.db.Exec(ctx, query, student.Id, student.IsActive)
	if err != nil {
		return "", err
	}
	return student.Id, nil
}

func (s *studentRepo) Delete(ctx context.Context, id string) error {
	query := `
	DELETE
	FROM
		students
	WHERE 
		id = $1 `

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *studentRepo) GetAll(ctx context.Context, req models.GetAllStudentsRequest) (models.GetAllStudentsResponse, error) {
	resp := models.GetAllStudentsResponse{}
	filter := ""
	offest := (req.Page - 1) * req.Limit

	if req.Search != "" {
		filter = ` AND first_name ILIKE '%` + req.Search + `%' `
	}

	query := `
	SELECT
		id,
		first_name,
		last_name,
		age,
		external_id,
		phone,
		mail,
		TO_CHAR(created_at,'YYYY-MM-DD HH:MM:SS'),
		TO_CHAR(updated_at,'YYYY-MM-DD HH:MM:SS'),
		is_active
	FROM
		students
	WHERE TRUE ` + filter + `
	OFFSET
		$1
	LIMIT
		$2;`

	rows, err := s.db.Query(ctx, query, offest, req.Limit)
	if err != nil {
		return resp, err
	}
	for rows.Next() {
		var (
			student                                                 models.GetStudent
			firstName, lastName, externalId, phone, mail, updatedAt sql.NullString
		)
		if err := rows.Scan(
			&student.Id,
			&firstName,
			&lastName,
			&student.Age,
			&externalId,
			&phone,
			&mail,
			&student.CreatedAt,
			&updatedAt,
			&student.IsActive); err != nil {
			return resp, err
		}
		student.FirstName = pkg.NullStringToString(firstName)
		student.LastName = pkg.NullStringToString(lastName)
		student.ExternalId = pkg.NullStringToString(externalId)
		student.Phone = pkg.NullStringToString(phone)
		student.Email = pkg.NullStringToString(mail)
		student.UpdatedAt = pkg.NullStringToString(updatedAt)

		resp.Students = append(resp.Students, student)
	}

	err = s.db.QueryRow(ctx, `SELECT count(*) from students WHERE TRUE `+filter+``).Scan(&resp.Count)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *studentRepo) GetStudent(ctx context.Context, id string) (models.GetStudent, error) {

	query := `
	SELECT
		id,
		first_name,
		last_name,
		age,
		external_id,
		phone,
		mail,
		TO_CHAR(created_at,'YYYY-MM-DD HH:MM:SS'),
		TO_CHAR(updated_at,'YYYY-MM-DD HH:MM:SS'),
		is_active
	FROM
		students
	WHERE
		id = $1;`

	row := s.db.QueryRow(ctx, query, id)

	var (
		student                                                 models.GetStudent
		firstName, lastName, externalId, phone, mail, updatedAt sql.NullString
	)

	err := row.Scan(&student.Id, &firstName, &lastName, &student.Age, &externalId, &phone, &mail, &student.CreatedAt, &updatedAt, &student.IsActive)

	student.FirstName = pkg.NullStringToString(firstName)
	student.LastName = pkg.NullStringToString(lastName)
	student.ExternalId = pkg.NullStringToString(externalId)
	student.Phone = pkg.NullStringToString(phone)
	student.Email = pkg.NullStringToString(mail)
	student.UpdatedAt = pkg.NullStringToString(updatedAt)

	if err != nil {
		return student, err
	}

	return student, nil
}

func (s *studentRepo) CheckStudentLesson(ctx context.Context, id string) (models.CheckLessonStudent, error) {

	query := `
	SELECT
		st.first_name || ' ' || st.last_name AS student_name,
		st.age,
		sb.name AS subject_name,
		ts.first_name || ' ' || ts.last_name AS teacher_name,
		tt.room_name,
		tt.to_date
	FROM
		students st
	INNER JOIN
		time_table tt
	ON
		st.id = tt.student_id
	INNER JOIN
		subjects sb
	ON
		sb.id = tt.subject_id
	INNER JOIN
		teachers ts
	ON
		ts.id = tt.teacher_id
	WHERE 
		st.id = $1;`

	row := s.db.QueryRow(ctx, query, id)

	var (
		checkStudent                                    models.CheckLessonStudent
		studentName, subjectName, teacherName, roomName sql.NullString
		savedTime time.Time
	)

	err := row.Scan(&studentName, &checkStudent.StudentAge, &subjectName, &teacherName, &roomName, &savedTime)

	if err != nil {
		return models.CheckLessonStudent{}, err
	}

	checkStudent.StudentName = pkg.NullStringToString(studentName)
	checkStudent.SubjectName = pkg.NullStringToString(subjectName)
	checkStudent.TeacherName = pkg.NullStringToString(teacherName)
	checkStudent.RoomName = pkg.NullStringToString(roomName)


	currentTime := time.Now()

	savedTimePart := savedTime.Format("15:04:05")
	currentTimePart := currentTime.Format("15:04:05")

	changeFormatSavedTime, err := time.Parse("15:04:05", savedTimePart)

	if err != nil {
		return models.CheckLessonStudent{}, err
	}

	changeFormatCurrentTime, err := time.Parse("15:04:05", currentTimePart)

	if err != nil {
		return models.CheckLessonStudent{}, err
	}

	difference := changeFormatSavedTime.Sub(changeFormatCurrentTime)

	checkStudent.TimeLeft = difference.Minutes()

	return checkStudent, nil
}


func (s *studentRepo) GetAllStudentsAttandenceReport(ctx context.Context, req models.GetAllStudentsAttandenceReportRequest) (models.GetAllStudentsAttandenceReportResponse, error) {
	resp := models.GetAllStudentsAttandenceReportResponse{}
	filter := ""
	offest := (req.Page - 1) * req.Limit

	if req.StudentId != "" {
		filter = ` AND s.id =` + req.StudentId + ` `
	}

	if req.TeacherId != "" {
		filter += ` AND t.id =` + req.TeacherId + ` `
	}

	if req.StartDate != "" && req.EndDate != "" {
		filter += ` AND tt.from_date BETWEEN '` + req.StartDate + `' AND '` + req.EndDate + `' `
	}

	// 	1. Student name,
	// 	2. student createdAt,
	// 	3. o’qituvchi name,
	// 	4. studying_time,
	// 	5. avg_studying_time,
	query := `
	SELECT
        s.id,
        s.first_name || ' ' || s.last_name AS student_name,
        TO_CHAR(s.created_at,'YYYY-MM-DD HH24:MI:SS'),
        t.first_name || ' ' || t.last_name AS teacher_name,
        EXTRACT(EPOCH FROM (tt.to_date - tt.from_date)) / 3600 AS studing_time
                
    FROM 
        time_table tt
        JOIN students s on tt.student_id = s.id
        JOIN teachers t on tt.teacher_id = t.id
        WHERE
		TRUE ` + filter + `
	OFFSET $1 LIMIT $2;`

	rows, err := s.db.Query(ctx, query, offest, req.Limit)
	if err != nil {
		return resp, err
	}
	studentAttandence := models.StudentAttandenceReport{}
	for rows.Next() {

		if err := rows.Scan(
			&studentAttandence.StudentId,
			&studentAttandence.StudentName,
			&studentAttandence.StudentCreatedAt,
			&studentAttandence.TeacherName,
			&studentAttandence.StudyTime); err != nil {
			return resp, err
		}

		resp.Students = append(resp.Students, studentAttandence)
	}

	err = s.db.QueryRow(ctx, `SELECT COUNT(*) from time_table time_table tt
	JOIN students s on tt.student_id = s.id
	JOIN teachers t on tt.teacher_id = t.id WHERE TRUE `+filter+``).Scan(&resp.Count)
	if err != nil {
		return resp, err
	}
	return resp, nil
}



func (s *studentRepo) UploadImage(ctx context.Context, path models.UploadStudentImage) error {
	query := `
	UPDATE
		students
	SET
		photo = $2, updated_at = NOW()
	WHERE 
		id = $1; `

	_, err := s.db.Exec(ctx, query, path.Id, path.Path)
	if err != nil {
		return err
	}
	return nil
}