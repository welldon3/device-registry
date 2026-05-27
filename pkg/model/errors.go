package model

type NotFoundError struct{ ID string }

func (e *NotFoundError) Error() string { return "device not found: " + e.ID }

type ConflictError struct{ ID string }

func (e *ConflictError) Error() string { return "device already exists: " + e.ID }

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }