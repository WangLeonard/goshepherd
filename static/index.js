const $ = layui.jquery;
init();

function submitForm(data) {
	if ($("#openBtn").hasClass("layui-btn-disabled")) {
		return false;
	}
	$("#openBtn").addClass("layui-btn-disabled");

	toolType = $("#tool").val();
	prjName = $("#projectName").val();
	path1 = $("#path1").val();
	path2 = $("#path2").val() ==="dummy-placeholder" ? "" : $("#path2").val()
	$.ajax({
		type: "get",
		url: "http://"+getCurrUrl()+"/api",
		data: {
			op: "add",
			tool: toolType,
			name: prjName,
			path1: path1,
			path2: path2
		},
		success: function (data) {
			if (data === "") {
				alertCheckGoShepherd();
				return false;
			}

			var port = Number(data);
			if (isNaN(port)) {
				layer.alert("err", {
					type: 0,
					title: `Warning`,
					content: data
				})
				return false;
			}

			pathNode = '<td><p>'+path1+'</p>'
			if (path2 !== "") {
				pathNode += '<p>'+path2+'</p>'
			}
			pathNode += '</td>'

			$("#tableBody").append('<tr>\n' +
				'<td>' + prjName + '</td>\n' +
				'<td><a style="color:#009688" href="http://'+ getCurrIP() + ':'+port+'" target="_blank">http://'+ getCurrIP() + ':'+port+'</a></td>\n' +
				pathNode+
				'<td>\n' +
				' <button type="button" port='+port+' class="delBtn layui-btn layui-btn-sm layui-btn-danger">\n' +
				' <i class="layui-icon">&#xe640;</i>\n' +
				' </button>\n' +
				'</td>\n' +
				'</tr>');

			return false;
		},
		error: function (data) {
			alertCheckGoShepherd();
			return false;
		}
	});

	$("#openBtn").removeClass("layui-btn-disabled");
	return false;
}

layui.use('form', function () {
	const form = layui.form;
	form.on('select(toolSelect)', function (data) {
		const path2Input = $("#path2Input");
		if (data.value === "2") {
			if (path2Input.hasClass("hidden") === true) {
				path2Input.removeClass("hidden");
				$("#path2").attr("lay-verify", "required").val("");
				form.render();
			}
		} else {
			if (path2Input.hasClass("hidden") === false) {
				path2Input.addClass("hidden");
				$("#path2").attr("lay-verify", "").val("dummy-placeholder");
				form.render();
			}
		}
	});

	// submit form
	form.on('submit(myForm)', submitForm);

	// delete item
	$("#tableBody").on("click",".delBtn", function () {
		$.ajax({
			type: "get",
			url: "http://"+ getCurrUrl() + "/api",
			data: {
				op: "rmv",
				port: $(this).attr("port"),
			}
		})
		$(this).parent().parent().remove();
	});
});

layui.use('upload', function(){
	var upload = layui.upload;
	//执行实例
	var uploadInst = upload.render({
		elem: '#test1' //绑定元素
		,url: '/upload/' //上传接口
		,multiple: true
		,data:{  proName: function(){
				return $('#projectName').val();
			}}
		,accept: 'file'
		,done: function(res){
			console.log(res["Path"]);
			//上传完毕回调
			$("#path1").attr("value",res["Path"]);
			submitForm();

		}
		,error: function(){
			console.log("err")
			//请求异常回调
		}
	});
});


function alertCheckGoShepherd() {
	var layer = layui.layer;
	layer.alert("err", {
		type: 0,
		title: `Warning`,
		content: `Oops...Seems that GoShepherd is not working well...Please check it!`
	})
}

function getCurrUrl(){
	return window.location.host;
}

function getCurrIP(){
	return getCurrUrl().split(":")[0];
}

function init(){
	$.ajax({
		type: "get",
		url: "http://"+getCurrUrl()+"/api",
		data: {
			op: "get",
		},
		success: function (data) {
			if (data === "") {
				alertCheckGoShepherd();
				return false;
			}
			var jsonArray = eval(data)
			jsonArray.forEach((item) => {
				console.log(item);
				path1 = item["Path1"]
				path2 = item["Path2"]
				prjName = item["Name"]
				port = item["Port"]

				pathNode = '<td><p>'+path1+'</p>'
				if (path2 !== "") {
					pathNode += '<p>'+path2+'</p>'
				}
				pathNode += '</td>'

				$("#tableBody").append('<tr>\n' +
					'<td>' + prjName + '</td>\n' +
					'<td><a style="color:#009688" href="http://'+ getCurrIP() + ':'+port+'" target="_blank">http://'+ getCurrIP() + ':'+port+'</a></td>\n' +
					pathNode+
					'<td>\n' +
					' <button type="button" port='+port+' class="delBtn layui-btn layui-btn-sm layui-btn-danger">\n' +
					' <i class="layui-icon">&#xe640;</i>\n' +
					' </button>\n' +
					'</td>\n' +
					'</tr>');
			});
			return;
		},
		error: function (data) {
			alertCheckGoShepherd();
			return false;
		}
	});
}
