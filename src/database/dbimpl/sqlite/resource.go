package sqlite

import "fsrv/src/database/entities"

func (sqlite *SQLiteDB) CreateResource(resource *entities.Resource) error {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}

	//insert resource with flags
	stmt := tx.Stmt(sqlite.qm.InsResourceData)
	_, err = stmt.Exec(resource.ID, resource.Flags)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//insert permissions
	err = sqlite.createResourcePermissions(tx, resource)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//commit transaction results
	commitOrPanic(tx)
	return nil
}

func (sqlite *SQLiteDB) DeleteResource(id string) error {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}

	//delete associated permissions
	stmt := tx.Stmt(sqlite.qm.DelPermissionByResourceID)
	_, err = stmt.Exec(id)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//delete underlying resource
	stmt = tx.Stmt(sqlite.qm.DelResourceByID)
	_, err = stmt.Exec(id)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//commit
	commitOrPanic(tx)
	return nil
}

func (sqlite *SQLiteDB) GetResources(pageSize int, offset int) ([]*entities.Resource, error) {
	resourceIDs, err := sqlite.GetResourceIDs(pageSize, offset)
	if err != nil {
		return nil, nil
	}
	resources := make([]*entities.Resource, 0, len(resourceIDs))

	for i, id := range resourceIDs {
		resources[i], err = sqlite.GetResourceData(id)
		if err != nil {
			return resources, err
		}
	}

	return resources, nil
}

func (sqlite *SQLiteDB) GetResourceIDs(pageSize int, offset int) ([]string, error) {
	var resourceIDs []string
	var id string
	rows, err := sqlite.db.Query("SELECT resourceid FROM Resources LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return resourceIDs, err
	}

	for rows.Next() {
		err = rows.Scan(&id)
		resourceIDs = append(resourceIDs, id)
		if err != nil {
			return resourceIDs, err
		}
	}

	return resourceIDs, nil
}

func (sqlite *SQLiteDB) GetResourceData(resourceid string) (*entities.Resource, error) {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return nil, err
	}

	var res entities.Resource

	//get flags
	stmt := tx.Stmt(sqlite.qm.GetResourceFlagsByID)
	row := stmt.QueryRow(resourceid)

	err = row.Scan(&res.Flags)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	//get permission iterator
	iter, roleperm, err := sqlite.getResourceRolePermIter(tx, resourceid)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	//get permissions
	for iter() == nil {
		key := entities.ResourceOperationAccess{
			ID:   roleperm.Role.ID,
			Type: roleperm.Perm.TypeRWMD,
		}
		res.OperationNodes[key] = roleperm.Perm.Status
	}

	_ = tx.Commit()

	return &res, nil
}
