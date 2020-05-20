package main

type Page struct {
	NameTitle           string //Frontend Name
	Name                string //Template Name
	URLPattern          string //URL Pattern
	Internal            bool   //Require User to be Logged In
	RequiresPermissions string //Permissions Required to View Page
}
