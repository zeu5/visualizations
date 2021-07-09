package crime

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zeu5/visualizations/models/crime"
	"github.com/zeu5/visualizations/server/common"
)

func Tables(c *gin.Context) {
	yearS := c.Query("year")
	if yearS != "" {
		year, err := strconv.Atoi(yearS)
		if err != nil {
			c.Error(errors.New("bad year parameter"))
			c.JSON(http.StatusBadRequest, common.Response{
				Error: "invalid error parameter",
			})
			return
		}
		tables, err := crime.TablesByYear(year, true)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, &common.Response{
				Error: "failed to fetch data from database",
			})
			return
		}
		c.JSON(http.StatusOK, &common.Response{
			Data: tables,
		})
	} else {
		tables, err := crime.AllTables(true)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, &common.Response{
				Error: "failed to fetch data from database",
			})
			return
		}
		c.JSON(http.StatusOK, &common.Response{
			Data: tables,
		})
	}
}

func Table(c *gin.Context) {
	dID := c.Param("id")
	if dID == "" {
		c.Error(errors.New("no id parameter"))
		c.JSON(http.StatusBadRequest, common.Response{
			Error: "no id parameter",
		})
		return
	}
	id, err := strconv.ParseUint(dID, 10, 64)
	if err != nil {
		c.Error(errors.New("bad id parameter"))
		c.JSON(http.StatusBadRequest, common.Response{
			Error: "invalid id parameter",
		})
		return
	}
	table, err := crime.TableByID(id)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, &common.Response{
			Error: "failed to fetch data from database",
		})
		return
	}
	c.JSON(http.StatusOK, common.Response{
		Data: table,
	})
}

func Initialize(router *gin.RouterGroup) {
	router.GET("/tables", Tables)
	router.GET("/tables/:id", Table)
}
