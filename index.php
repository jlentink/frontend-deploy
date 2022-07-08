<!DOCTYPE html>
<html>
<head>
	<title>Branches deployed</title>
	<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
	<style>
		body {
			height: 100%;
			background-color: #e7e7e7 !important;
			background-color: white;
			font-family: sans-serif;
		}
		.container-fluid {
			margin-top: 80px;
		}
		footer {
			margin: 20px;
		}
	</style>
</head>
<body>
	<div class="container-fluid container-table">
		<div class="row vertical-center-row">
        	<div class="col-md-6">
        		<h1>
        		Welcome to<br />
        		<a href="//<?php print $_SERVER['SERVER_NAME']; ?>"><?php print $_SERVER['SERVER_NAME'];?></a>
        		</h1>
				This file functions as an overview page for all the different builds.

			</div>
        	<div class="col-md-6">
				<h2>Builds</h2>
				<table class="table table-striped">
		           <thead>
		              <tr>
		                <th>#</th>
		                <th>Branch</th>
		                <th>Date modified</th>
		              </tr>
		            </thead>
		            <tbody>
						<?php
							$entries = scandir(".");
							$index = 1;
							foreach ($entries as $key => $entry) {
								if(is_dir($entry) && substr($entry, 0, 1) != ".") {
									?>
					            	<tr>
					            		<td><?php print $index; ?></td>
					            		<td><?php print "<a href=\"" . $entry . "\">" . $entry . "</a>"; ?></td>
					            		<?php if(is_file($entry ."/deploy.json")){ ?>
					            		<td><?php
					            		    $data = json_decode(file_get_contents($entry ."/deploy.json"));
					            		    print date ("d.m.Y H:i e", $data->deployDate);
					            		?></td>
					            		<?php } else{ ?>
					            		<td><?php print date ("d.m.Y H:i e", filemtime(__DIR__ . "/" . $entry)); ?></td>
					            		<?php } ?>
					            	</tr>
									<?php
									$index++;
								}
							}
						?>
		            </tbody>
				</table>

			</div>
		</div>
	</div>
	<footer>
		<hr>
		Date: <?php print date ("F d Y H:i:s e.", filemtime(__FILE__)); ?>
	</footer>
</body>
</html>