!function(){"use strict";var i=window.location,o=window.document,p=o.getElementById("plausible"),u=p.getAttribute("data-api")||(m=(m=p).src.split("/"),l=m[0],m=m[2],l+"//"+m+"/api/event");function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return a=t,(n="localStorage flag")&&console.warn("Ignoring Event: "+n),void(a&&a.callback&&a.callback())}catch(e){}var a,n={},r=(n.n=e,n.u=i.href,n.d=p.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),new XMLHttpRequest);r.open("POST",u,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function r(){a!==i.pathname&&(a=i.pathname,e("pageview"))}var s,l=window.history;function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}l.pushState&&(s=l.pushState,l.pushState=function(){s.apply(this,arguments),r()},window.addEventListener("popstate",r)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){a||"visible"!==o.visibilityState||r()}):r();var f=1;function v(e){if("auxclick"!==e.type||e.button===f){var t,a,n=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),r=n&&n.href&&n.href.split("?")[0];if(!function e(t,a){if(!t||y<a)return!1;if(L(t))return!0;return e(t.parentNode,a+1)}(n,0))return(t=n)&&t.href&&t.host&&t.host!==i.host?d(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):(t=r)&&(a=t.split(".").pop(),h.some(function(e){return e===a}))?d(e,n,{name:"File Download",props:{url:r}}):void 0}}function d(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((n={props:a.props}).revenue=a.revenue,plausible(a.name,n)):((n={props:a.props,callback:i}).revenue=a.revenue,plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",v),o.addEventListener("auxclick",v);var m=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],g=p.getAttribute("file-types"),b=p.getAttribute("add-file-types"),h=g&&g.split(",")||b&&b.split(",").concat(m)||m;function w(e){var e=L(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i,o=a.item(n),p=o.match(/plausible-event-(.+)(=|--)(.+)/),p=(p&&(r=p[1],i=p[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i),o.match(/plausible-revenue-(.+)(=|--)(.+)/));p&&(r=p[1],i=p[3],t.revenue[r]=i)}return t}var y=3;function k(e){if("auxclick"!==e.type||e.button===f){for(var t,a,n,r,i=e.target,o=0;o<=y&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;c(i)&&(t=i),L(i)&&(a=i),i=i.parentNode}a&&(r=w(a),t?(r.props.url=t.href,d(e,t,r)):((e={}).props=r.props,e.revenue=r.revenue,plausible(r.name,e)))}}function L(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,a=e.target,n=w(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),(e={props:n.props,callback:r}).revenue=n.revenue,plausible(n.name,e))}),o.addEventListener("click",k),o.addEventListener("auxclick",k)}();